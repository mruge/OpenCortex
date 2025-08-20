package models

type Request struct {
	Operation     string      `json:"operation"`
	CorrelationID string      `json:"correlation_id"`
	Query         QueryData   `json:"query"`
	Enrich        []string    `json:"enrich,omitempty"`
	Limit         int         `json:"limit,omitempty"`
}

type QueryData struct {
	Cypher    string    `json:"cypher,omitempty"`
	Text      string    `json:"text,omitempty"`
	Embedding []float32 `json:"embedding,omitempty"`
	NodeIDs   []string  `json:"node_ids,omitempty"`
}

const (
	OperationTraverse = "traverse"
	OperationSearch   = "search"
	OperationEnrich   = "enrich"
)