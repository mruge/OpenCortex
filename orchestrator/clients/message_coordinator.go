package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"orchestrator/models"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RedisMessageCoordinator handles communication with other services via Redis pub/sub
type RedisMessageCoordinator struct {
	client         *redis.Client
	dataConfig     ServiceChannelConfig
	aiConfig       ServiceChannelConfig  
	execConfig     ServiceChannelConfig
	defaultTimeout time.Duration
	subscribers    map[string]*redis.PubSub
	responseWaiters map[string]chan *models.ServiceResponse
	mutex          sync.RWMutex
	logger         *logrus.Logger
}

// ServiceChannelConfig defines channel configuration for a service
type ServiceChannelConfig struct {
	RequestChannel  string
	ResponseChannel string
	Timeout         time.Duration
}

// NewRedisMessageCoordinator creates a new Redis-based message coordinator
func NewRedisMessageCoordinator(client *redis.Client, dataConfig, aiConfig, execConfig ServiceChannelConfig, defaultTimeout time.Duration) *RedisMessageCoordinator {
	mc := &RedisMessageCoordinator{
		client:          client,
		dataConfig:      dataConfig,
		aiConfig:        aiConfig,
		execConfig:      execConfig,
		defaultTimeout:  defaultTimeout,
		subscribers:     make(map[string]*redis.PubSub),
		responseWaiters: make(map[string]chan *models.ServiceResponse),
		logger:          logrus.New(),
	}

	// Start response listeners
	mc.startResponseListener(dataConfig.ResponseChannel)
	mc.startResponseListener(aiConfig.ResponseChannel)
	mc.startResponseListener(execConfig.ResponseChannel)

	return mc
}

// SendDataRequest sends a request to the data service
func (mc *RedisMessageCoordinator) SendDataRequest(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error) {
	return mc.sendServiceRequest(ctx, request, mc.dataConfig)
}

// SendAIRequest sends a request to the AI service
func (mc *RedisMessageCoordinator) SendAIRequest(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error) {
	return mc.sendServiceRequest(ctx, request, mc.aiConfig)
}

// SendExecRequest sends a request to the execution service
func (mc *RedisMessageCoordinator) SendExecRequest(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error) {
	return mc.sendServiceRequest(ctx, request, mc.execConfig)
}

// sendServiceRequest is the generic implementation for sending requests
func (mc *RedisMessageCoordinator) sendServiceRequest(ctx context.Context, request *models.ServiceRequest, config ServiceChannelConfig) (*models.ServiceResponse, error) {
	// Generate unique correlation ID if not provided
	if request.CorrelationID == "" {
		request.CorrelationID = uuid.New().String()
	}

	// Set service in request
	request.Service = mc.getServiceNameFromConfig(config)

	mc.logger.WithFields(logrus.Fields{
		"correlation_id": request.CorrelationID,
		"service":        request.Service,
		"operation":      request.Operation,
		"channel":        config.RequestChannel,
	}).Info("Sending service request")

	// Create response waiter
	responseChan := make(chan *models.ServiceResponse, 1)
	mc.mutex.Lock()
	mc.responseWaiters[request.CorrelationID] = responseChan
	mc.mutex.Unlock()

	// Cleanup waiter on return
	defer func() {
		mc.mutex.Lock()
		delete(mc.responseWaiters, request.CorrelationID)
		mc.mutex.Unlock()
		close(responseChan)
	}()

	// Marshal request
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Publish request
	err = mc.client.Publish(ctx, config.RequestChannel, requestData).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	// Wait for response with timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = mc.defaultTimeout
	}

	select {
	case response := <-responseChan:
		mc.logger.WithFields(logrus.Fields{
			"correlation_id": request.CorrelationID,
			"service":        request.Service,
			"success":        response.Success,
		}).Info("Received service response")
		return response, nil

	case <-time.After(timeout):
		return nil, fmt.Errorf("request timeout after %v", timeout)

	case <-ctx.Done():
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	}
}

// startResponseListener starts listening for responses on a channel
func (mc *RedisMessageCoordinator) startResponseListener(channel string) {
	pubsub := mc.client.Subscribe(context.Background(), channel)
	mc.subscribers[channel] = pubsub

	go mc.responseListener(channel, pubsub)

	mc.logger.WithField("channel", channel).Info("Started response listener")
}

// responseListener processes incoming responses
func (mc *RedisMessageCoordinator) responseListener(channel string, pubsub *redis.PubSub) {
	defer pubsub.Close()

	ch := pubsub.Channel()
	
	for msg := range ch {
		var response models.ServiceResponse
		if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
			mc.logger.WithError(err).WithFields(logrus.Fields{
				"channel": channel,
				"payload": msg.Payload,
			}).Error("Failed to unmarshal service response")
			continue
		}

		// Route response to waiting caller
		mc.mutex.RLock()
		responseChan, exists := mc.responseWaiters[response.CorrelationID]
		mc.mutex.RUnlock()

		if exists {
			select {
			case responseChan <- &response:
				// Response delivered
			default:
				// Channel full or closed, log warning
				mc.logger.WithField("correlation_id", response.CorrelationID).Warn("Failed to deliver response - channel full or closed")
			}
		} else {
			// No waiter for this response - might be expired
			mc.logger.WithField("correlation_id", response.CorrelationID).Debug("Received response with no waiting caller")
		}
	}

	mc.logger.WithField("channel", channel).Info("Response listener stopped")
}

// getServiceNameFromConfig determines service name from channel config
func (mc *RedisMessageCoordinator) getServiceNameFromConfig(config ServiceChannelConfig) string {
	switch config.RequestChannel {
	case mc.dataConfig.RequestChannel:
		return "data"
	case mc.aiConfig.RequestChannel:
		return "ai"
	case mc.execConfig.RequestChannel:
		return "exec"
	default:
		return "unknown"
	}
}

// SendBroadcastRequest sends a request that expects multiple responses
func (mc *RedisMessageCoordinator) SendBroadcastRequest(ctx context.Context, request *models.ServiceRequest, expectedResponses int, timeout time.Duration) ([]*models.ServiceResponse, error) {
	// Generate unique correlation ID
	if request.CorrelationID == "" {
		request.CorrelationID = uuid.New().String()
	}

	mc.logger.WithFields(logrus.Fields{
		"correlation_id":      request.CorrelationID,
		"expected_responses":  expectedResponses,
		"timeout":            timeout,
	}).Info("Sending broadcast request")

	// Create response collector
	responses := make([]*models.ServiceResponse, 0, expectedResponses)
	responseChan := make(chan *models.ServiceResponse, expectedResponses)
	
	mc.mutex.Lock()
	mc.responseWaiters[request.CorrelationID] = responseChan
	mc.mutex.Unlock()

	// Cleanup waiter on return
	defer func() {
		mc.mutex.Lock()
		delete(mc.responseWaiters, request.CorrelationID)
		mc.mutex.Unlock()
		close(responseChan)
	}()

	// Marshal and send request to all services
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal broadcast request: %w", err)
	}

	// Send to all service channels
	channels := []string{
		mc.dataConfig.RequestChannel,
		mc.aiConfig.RequestChannel,
		mc.execConfig.RequestChannel,
	}

	for _, channel := range channels {
		if err := mc.client.Publish(ctx, channel, requestData).Err(); err != nil {
			mc.logger.WithError(err).WithField("channel", channel).Error("Failed to publish broadcast request")
		}
	}

	// Collect responses
	timeoutChan := time.After(timeout)
	
	for len(responses) < expectedResponses {
		select {
		case response := <-responseChan:
			responses = append(responses, response)
			
		case <-timeoutChan:
			mc.logger.WithFields(logrus.Fields{
				"correlation_id": request.CorrelationID,
				"received":       len(responses),
				"expected":       expectedResponses,
			}).Warn("Broadcast request timeout - returning partial results")
			return responses, fmt.Errorf("broadcast timeout: received %d/%d responses", len(responses), expectedResponses)
			
		case <-ctx.Done():
			return responses, fmt.Errorf("broadcast cancelled: %w", ctx.Err())
		}
	}

	mc.logger.WithFields(logrus.Fields{
		"correlation_id": request.CorrelationID,
		"responses":      len(responses),
	}).Info("Broadcast request completed")

	return responses, nil
}

// SendAsyncRequest sends a request without waiting for response
func (mc *RedisMessageCoordinator) SendAsyncRequest(ctx context.Context, request *models.ServiceRequest, config ServiceChannelConfig) error {
	// Generate correlation ID if not provided
	if request.CorrelationID == "" {
		request.CorrelationID = uuid.New().String()
	}

	request.Service = mc.getServiceNameFromConfig(config)

	mc.logger.WithFields(logrus.Fields{
		"correlation_id": request.CorrelationID,
		"service":        request.Service,
		"operation":      request.Operation,
	}).Info("Sending async service request")

	// Marshal request
	requestData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal async request: %w", err)
	}

	// Publish request
	return mc.client.Publish(ctx, config.RequestChannel, requestData).Err()
}

// GetStats returns statistics about the message coordinator
func (mc *RedisMessageCoordinator) GetStats() map[string]interface{} {
	mc.mutex.RLock()
	pendingRequests := len(mc.responseWaiters)
	mc.mutex.RUnlock()

	return map[string]interface{}{
		"pending_requests":  pendingRequests,
		"active_subscribers": len(mc.subscribers),
		"data_channel":      mc.dataConfig.RequestChannel,
		"ai_channel":        mc.aiConfig.RequestChannel,
		"exec_channel":      mc.execConfig.RequestChannel,
	}
}

// Close stops all listeners and cleans up resources
func (mc *RedisMessageCoordinator) Close() error {
	mc.logger.Info("Closing message coordinator")

	// Close all subscribers
	for channel, pubsub := range mc.subscribers {
		if err := pubsub.Close(); err != nil {
			mc.logger.WithError(err).WithField("channel", channel).Error("Error closing subscriber")
		}
	}

	// Clear response waiters
	mc.mutex.Lock()
	for corrID, ch := range mc.responseWaiters {
		close(ch)
		delete(mc.responseWaiters, corrID)
	}
	mc.mutex.Unlock()

	return nil
}

// HealthCheck verifies connectivity to Redis
func (mc *RedisMessageCoordinator) HealthCheck(ctx context.Context) error {
	return mc.client.Ping(ctx).Err()
}