package clients

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/sirupsen/logrus"
)

type MongoClient struct {
	client   *mongo.Client
	database string
}

func NewMongoClient(url, database string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"url":      url,
		"database": database,
	}).Info("MongoDB client connected")

	return &MongoClient{
		client:   client,
		database: database,
	}, nil
}

func (m *MongoClient) GetEnrichmentData(ctx context.Context, nodeIDs []string) (map[string]interface{}, error) {
	if len(nodeIDs) == 0 {
		return make(map[string]interface{}), nil
	}

	collection := m.client.Database(m.database).Collection("enrichment")
	
	filter := bson.M{"node_id": bson.M{"$in": nodeIDs}}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	enrichmentData := make(map[string]interface{})
	
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			logrus.WithError(err).Warn("Failed to decode enrichment document")
			continue
		}

		if nodeID, ok := doc["node_id"].(string); ok {
			delete(doc, "_id")
			delete(doc, "node_id")
			enrichmentData[nodeID] = doc
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"requested": len(nodeIDs),
		"found":     len(enrichmentData),
	}).Debug("MongoDB enrichment data retrieved")

	return enrichmentData, nil
}

func (m *MongoClient) GetMetadataByNodeID(ctx context.Context, nodeID string) (map[string]interface{}, error) {
	collection := m.client.Database(m.database).Collection("metadata")
	
	filter := bson.M{"node_id": nodeID}
	
	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	delete(result, "_id")
	delete(result, "node_id")

	return result, nil
}

func (m *MongoClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return m.client.Disconnect(ctx)
}