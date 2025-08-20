package clients

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisClient struct {
	client     *redis.Client
	requestCh  string
	responseCh string
}

func NewRedisClient(url, requestCh, responseCh string) (*RedisClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"url":        url,
		"request_ch": requestCh,
		"response_ch": responseCh,
	}).Info("Exec Agent Redis client connected")

	return &RedisClient{
		client:     client,
		requestCh:  requestCh,
		responseCh: responseCh,
	}, nil
}

func (r *RedisClient) Listen(ctx context.Context, handler func([]byte) []byte) error {
	pubsub := r.client.Subscribe(ctx, r.requestCh)
	defer pubsub.Close()

	ch := pubsub.Channel()

	logrus.WithField("channel", r.requestCh).Info("Starting Exec Agent listener")

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Exec Agent Redis listener shutting down")
			return ctx.Err()
		case msg := <-ch:
			if msg == nil {
				continue
			}

			logrus.WithFields(logrus.Fields{
				"channel": msg.Channel,
				"payload_size": len(msg.Payload),
			}).Debug("Received execution request")

			response := handler([]byte(msg.Payload))
			
			if err := r.Publish(ctx, r.responseCh, response); err != nil {
				logrus.WithError(err).Error("Failed to publish execution response")
			}
		}
	}
}

func (r *RedisClient) Publish(ctx context.Context, channel string, data []byte) error {
	return r.client.Publish(ctx, channel, data).Err()
}

func (r *RedisClient) PublishJSON(ctx context.Context, channel string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.Publish(ctx, channel, jsonData)
}

func (r *RedisClient) PublishToChannel(channel string, data interface{}) error {
	return r.PublishJSON(context.Background(), channel, data)
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient returns the underlying redis.Client for advanced usage
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}