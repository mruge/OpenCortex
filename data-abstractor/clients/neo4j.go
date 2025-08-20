package clients

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/sirupsen/logrus"
)

type Neo4jClient struct {
	driver neo4j.DriverWithContext
}

type GraphResult struct {
	Nodes         []Node         `json:"nodes"`
	Relationships []Relationship `json:"relationships"`
}

type Node struct {
	ID         string                 `json:"id"`
	Labels     []string               `json:"labels"`
	Properties map[string]interface{} `json:"properties"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Relationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	StartNode  string                 `json:"start_node"`
	EndNode    string                 `json:"end_node"`
	Properties map[string]interface{} `json:"properties"`
}

func NewNeo4jClient(url, username, password string) (*Neo4jClient, error) {
	driver, err := neo4j.NewDriverWithContext(
		url,
		neo4j.BasicAuth(username, password, ""),
		func(config *neo4j.Config) {
			config.MaxConnectionLifetime = 30 * time.Minute
			config.MaxConnectionPoolSize = 10
		},
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, err
	}

	logrus.WithField("url", url).Info("Neo4j client connected")

	return &Neo4jClient{driver: driver}, nil
}

func (n *Neo4jClient) ExecuteCypher(ctx context.Context, cypher string, params map[string]interface{}) (*GraphResult, error) {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypherResult, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		graphResult := &GraphResult{
			Nodes:         make([]Node, 0),
			Relationships: make([]Relationship, 0),
		}

		nodeMap := make(map[int64]bool)
		relMap := make(map[int64]bool)

		for cypherResult.Next(ctx) {
			record := cypherResult.Record()
			
			for _, value := range record.Values {
				switch v := value.(type) {
				case neo4j.Node:
					if !nodeMap[v.GetId()] {
						node := Node{
							ID:         string(rune(v.GetId())),
							Labels:     v.Labels,
							Properties: v.Props,
						}
						graphResult.Nodes = append(graphResult.Nodes, node)
						nodeMap[v.GetId()] = true
					}
				case neo4j.Relationship:
					if !relMap[v.GetId()] {
						rel := Relationship{
							ID:         string(rune(v.GetId())),
							Type:       v.Type,
							StartNode:  string(rune(v.StartId)),
							EndNode:    string(rune(v.EndId)),
							Properties: v.Props,
						}
						graphResult.Relationships = append(graphResult.Relationships, rel)
						relMap[v.GetId()] = true
					}
				case neo4j.Path:
					for _, node := range v.Nodes {
						if !nodeMap[node.GetId()] {
							n := Node{
								ID:         string(rune(node.GetId())),
								Labels:     node.Labels,
								Properties: node.Props,
							}
							graphResult.Nodes = append(graphResult.Nodes, n)
							nodeMap[node.GetId()] = true
						}
					}
					for _, rel := range v.Relationships {
						if !relMap[rel.GetId()] {
							r := Relationship{
								ID:         string(rune(rel.GetId())),
								Type:       rel.Type,
								StartNode:  string(rune(rel.StartId)),
								EndNode:    string(rune(rel.EndId)),
								Properties: rel.Props,
							}
							graphResult.Relationships = append(graphResult.Relationships, r)
							relMap[rel.GetId()] = true
						}
					}
				}
			}
		}

		return graphResult, cypherResult.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.(*GraphResult), nil
}

func (n *Neo4jClient) GetNodesByIds(ctx context.Context, nodeIds []string) (*GraphResult, error) {
	if len(nodeIds) == 0 {
		return &GraphResult{Nodes: []Node{}, Relationships: []Relationship{}}, nil
	}

	cypher := "MATCH (n) WHERE n.id IN $node_ids RETURN n"
	params := map[string]interface{}{
		"node_ids": nodeIds,
	}

	return n.ExecuteCypher(ctx, cypher, params)
}

func (n *Neo4jClient) Close() error {
	return n.driver.Close(context.Background())
}