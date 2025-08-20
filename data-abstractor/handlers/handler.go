package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"data-abstractor/clients"
	"data-abstractor/models"

	"github.com/sirupsen/logrus"
)

type DataHandler struct {
	neo4j   *clients.Neo4jClient
	mongo   *clients.MongoClient
	qdrant  *clients.QdrantClient
}

func NewDataHandler(neo4j *clients.Neo4jClient, mongo *clients.MongoClient, qdrant *clients.QdrantClient) *DataHandler {
	return &DataHandler{
		neo4j:  neo4j,
		mongo:  mongo,
		qdrant: qdrant,
	}
}

func (h *DataHandler) HandleRequest(ctx context.Context, data []byte) []byte {
	var req models.Request
	if err := json.Unmarshal(data, &req); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal request")
		response := models.NewErrorResponse("", "", fmt.Sprintf("Invalid request format: %v", err))
		responseData, _ := json.Marshal(response)
		return responseData
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id": req.CorrelationID,
		"operation":      req.Operation,
	}).Info("Processing request")

	var response *models.Response

	switch req.Operation {
	case models.OperationTraverse:
		response = h.handleTraverse(ctx, &req)
	case models.OperationSearch:
		response = h.handleSearch(ctx, &req)
	case models.OperationEnrich:
		response = h.handleEnrich(ctx, &req)
	default:
		response = models.NewErrorResponse(req.CorrelationID, req.Operation, fmt.Sprintf("Unknown operation: %s", req.Operation))
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal response")
		errorResponse := models.NewErrorResponse(req.CorrelationID, req.Operation, "Internal error")
		responseData, _ = json.Marshal(errorResponse)
	}

	return responseData
}

func (h *DataHandler) handleTraverse(ctx context.Context, req *models.Request) *models.Response {
	if req.Query.Cypher == "" {
		return models.NewErrorResponse(req.CorrelationID, req.Operation, "Cypher query is required for traverse operation")
	}

	graphResult, err := h.neo4j.ExecuteCypher(ctx, req.Query.Cypher, nil)
	if err != nil {
		logrus.WithError(err).Error("Neo4j query failed")
		return models.NewErrorResponse(req.CorrelationID, req.Operation, fmt.Sprintf("Query failed: %v", err))
	}

	graphData := h.convertNeo4jToGraphData(graphResult)

	if h.shouldEnrich(req.Enrich) {
		if err := h.enrichNodes(ctx, graphData.Nodes); err != nil {
			logrus.WithError(err).Warn("Enrichment failed")
		}
	}

	return models.NewSuccessResponse(req.CorrelationID, req.Operation, graphData)
}

func (h *DataHandler) handleSearch(ctx context.Context, req *models.Request) *models.Response {
	var searchResults []clients.SearchResult
	var err error

	if len(req.Query.Embedding) > 0 {
		limit := uint64(req.Limit)
		if limit == 0 {
			limit = 100
		}
		searchResults, err = h.qdrant.SearchSimilar(ctx, req.Query.Embedding, limit)
	} else if req.Query.Text != "" {
		limit := uint64(req.Limit)
		if limit == 0 {
			limit = 100
		}
		searchResults, err = h.qdrant.SearchText(ctx, req.Query.Text, limit)
	} else {
		return models.NewErrorResponse(req.CorrelationID, req.Operation, "Either text or embedding is required for search operation")
	}

	if err != nil {
		logrus.WithError(err).Error("Qdrant search failed")
		return models.NewErrorResponse(req.CorrelationID, req.Operation, fmt.Sprintf("Search failed: %v", err))
	}

	nodeIDs := make([]string, len(searchResults))
	scoreMap := make(map[string]float32)
	for i, result := range searchResults {
		nodeIDs[i] = result.NodeID
		scoreMap[result.NodeID] = result.Score
	}

	graphResult, err := h.neo4j.GetNodesByIds(ctx, nodeIDs)
	if err != nil {
		logrus.WithError(err).Error("Neo4j node retrieval failed")
		return models.NewErrorResponse(req.CorrelationID, req.Operation, fmt.Sprintf("Node retrieval failed: %v", err))
	}

	graphData := h.convertNeo4jToGraphData(graphResult)

	for i := range graphData.Nodes {
		if score, exists := scoreMap[graphData.Nodes[i].ID]; exists {
			graphData.Nodes[i].Score = score
		}
	}

	if h.shouldEnrich(req.Enrich) {
		if err := h.enrichNodes(ctx, graphData.Nodes); err != nil {
			logrus.WithError(err).Warn("Enrichment failed")
		}
	}

	return models.NewSuccessResponse(req.CorrelationID, req.Operation, graphData)
}

func (h *DataHandler) handleEnrich(ctx context.Context, req *models.Request) *models.Response {
	if len(req.Query.NodeIDs) == 0 {
		return models.NewErrorResponse(req.CorrelationID, req.Operation, "Node IDs are required for enrich operation")
	}

	graphResult, err := h.neo4j.GetNodesByIds(ctx, req.Query.NodeIDs)
	if err != nil {
		logrus.WithError(err).Error("Neo4j node retrieval failed")
		return models.NewErrorResponse(req.CorrelationID, req.Operation, fmt.Sprintf("Node retrieval failed: %v", err))
	}

	graphData := h.convertNeo4jToGraphData(graphResult)

	if err := h.enrichNodes(ctx, graphData.Nodes); err != nil {
		logrus.WithError(err).Error("Enrichment failed")
		return models.NewErrorResponse(req.CorrelationID, req.Operation, fmt.Sprintf("Enrichment failed: %v", err))
	}

	return models.NewSuccessResponse(req.CorrelationID, req.Operation, graphData)
}

func (h *DataHandler) convertNeo4jToGraphData(result *clients.GraphResult) *models.GraphData {
	nodes := make([]models.GraphNode, len(result.Nodes))
	for i, node := range result.Nodes {
		nodes[i] = models.GraphNode{
			ID:         node.ID,
			Labels:     node.Labels,
			Properties: node.Properties,
			Metadata:   node.Metadata,
		}
	}

	relationships := make([]models.GraphRelationship, len(result.Relationships))
	for i, rel := range result.Relationships {
		relationships[i] = models.GraphRelationship{
			ID:         rel.ID,
			Type:       rel.Type,
			StartNode:  rel.StartNode,
			EndNode:    rel.EndNode,
			Properties: rel.Properties,
		}
	}

	return &models.GraphData{
		Nodes:         nodes,
		Relationships: relationships,
	}
}

func (h *DataHandler) shouldEnrich(enrichFields []string) bool {
	return len(enrichFields) > 0 && (contains(enrichFields, "metadata") || contains(enrichFields, "all"))
}

func (h *DataHandler) enrichNodes(ctx context.Context, nodes []models.GraphNode) error {
	nodeIDs := make([]string, len(nodes))
	for i, node := range nodes {
		nodeIDs[i] = node.ID
	}

	enrichmentData, err := h.mongo.GetEnrichmentData(ctx, nodeIDs)
	if err != nil {
		return err
	}

	for i := range nodes {
		if metadata, exists := enrichmentData[nodes[i].ID]; exists {
			if nodes[i].Metadata == nil {
				nodes[i].Metadata = make(map[string]interface{})
			}
			for k, v := range metadata.(map[string]interface{}) {
				nodes[i].Metadata[k] = v
			}
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func stringToInt64(s string) int64 {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	return 0
}