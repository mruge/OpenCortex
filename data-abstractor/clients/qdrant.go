package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type QdrantClient struct {
	baseURL    string
	collection string
	httpClient *http.Client
}

type SearchResult struct {
	NodeID string  `json:"node_id"`
	Score  float32 `json:"score"`
}

type qdrantSearchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type qdrantSearchResponse struct {
	Result []struct {
		ID      interface{}            `json:"id"`
		Score   float32                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"result"`
}

func NewQdrantClient(url, collection string) (*QdrantClient, error) {
	client := &QdrantClient{
		baseURL:    url,
		collection: collection,
		httpClient: &http.Client{},
	}

	logrus.WithFields(logrus.Fields{
		"url":        url,
		"collection": collection,
	}).Info("Qdrant client connected")

	return client, nil
}

func (q *QdrantClient) SearchSimilar(ctx context.Context, vector []float32, limit uint64) ([]SearchResult, error) {
	searchReq := qdrantSearchRequest{
		Vector:      vector,
		Limit:       int(limit),
		WithPayload: true,
	}

	reqBody, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, q.collection)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant search failed with status %d", resp.StatusCode)
	}

	var searchResp qdrantSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(searchResp.Result))
	for _, point := range searchResp.Result {
		var nodeID string
		if point.Payload != nil {
			if nodeIDValue, exists := point.Payload["node_id"]; exists {
				if stringValue, ok := nodeIDValue.(string); ok {
					nodeID = stringValue
				}
			}
		}

		results = append(results, SearchResult{
			NodeID: nodeID,
			Score:  point.Score,
		})
	}

	logrus.WithFields(logrus.Fields{
		"collection": q.collection,
		"limit":      limit,
		"results":    len(results),
	}).Debug("Qdrant similarity search completed")

	return results, nil
}

func (q *QdrantClient) SearchText(ctx context.Context, text string, limit uint64) ([]SearchResult, error) {
	logrus.WithFields(logrus.Fields{
		"text":       text,
		"collection": q.collection,
	}).Warn("Text search requires embedding service - returning empty results")
	
	return []SearchResult{}, nil
}

func (q *QdrantClient) Close() error {
	return nil
}