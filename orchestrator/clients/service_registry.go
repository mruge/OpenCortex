package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// ServiceRegistry collects and manages capability announcements from services
type ServiceRegistry struct {
	redisClient      *redis.Client
	capabilities     map[string]*ServiceCapability
	lastSeen         map[string]time.Time
	mutex            sync.RWMutex
	logger           *logrus.Logger
	subscriber       *redis.PubSub
	stopChan         chan struct{}
	staleThreshold   time.Duration
}

// ServiceCapability represents a service's announced capabilities
type ServiceCapability struct {
	Component    string                 `json:"component"`
	Timestamp    string                 `json:"timestamp"`
	Trigger      string                 `json:"trigger"`
	Capabilities *ServiceCapabilities   `json:"capabilities"`
	LastUpdated  time.Time              `json:"last_updated"`
}

// ServiceCapabilities represents the complete capability information for a service
type ServiceCapabilities struct {
	Operations      []Operation      `json:"operations"`
	MessagePatterns MessagePatterns  `json:"message_patterns"`
}

// Operation represents a single operation that a service can perform
type Operation struct {
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	InputExample      interface{} `json:"input_example"`
	OutputExample     interface{} `json:"output_example"`
	RetrySafe         bool        `json:"retry_safe"`
	EstimatedDuration string      `json:"estimated_duration"`
}

// MessagePatterns defines how a service communicates via Redis
type MessagePatterns struct {
	RequestChannel   string `json:"request_channel"`
	ResponseChannel  string `json:"response_channel"`
	CorrelationField string `json:"correlation_field"`
}

// ServiceRegistryStats provides statistics about the registry
type ServiceRegistryStats struct {
	TotalServices     int                        `json:"total_services"`
	ActiveServices    int                        `json:"active_services"`
	StaleServices     int                        `json:"stale_services"`
	LastUpdate        time.Time                  `json:"last_update"`
	ServiceSummary    map[string]ServiceSummary  `json:"service_summary"`
}

// ServiceSummary provides a brief summary of a service's capabilities
type ServiceSummary struct {
	Component       string    `json:"component"`
	OperationCount  int       `json:"operation_count"`
	LastSeen        time.Time `json:"last_seen"`
	RequestChannel  string    `json:"request_channel"`
	ResponseChannel string    `json:"response_channel"`
	IsActive        bool      `json:"is_active"`
}

const (
	DefaultStaleThreshold = 15 * time.Minute
	AnnouncementChannel   = "service_capability_announcements"
	RefreshRequestChannel = "capability_refresh_request"
)

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(redisClient *redis.Client, staleThreshold time.Duration) *ServiceRegistry {
	if staleThreshold == 0 {
		staleThreshold = DefaultStaleThreshold
	}

	return &ServiceRegistry{
		redisClient:    redisClient,
		capabilities:   make(map[string]*ServiceCapability),
		lastSeen:       make(map[string]time.Time),
		logger:         logrus.WithField("component", "service_registry"),
		staleThreshold: staleThreshold,
		stopChan:       make(chan struct{}),
	}
}

// Start begins listening for capability announcements
func (sr *ServiceRegistry) Start(ctx context.Context) error {
	sr.logger.Info("Starting service registry")

	// Subscribe to capability announcements
	sr.subscriber = sr.redisClient.Subscribe(ctx, AnnouncementChannel)

	// Start announcement listener
	go sr.listenForAnnouncements(ctx)

	// Start cleanup routine for stale services
	go sr.cleanupStaleServices(ctx)

	// Request initial capability refresh from all services
	sr.requestCapabilityRefresh(ctx, "")

	return nil
}

// Stop stops the service registry
func (sr *ServiceRegistry) Stop() {
	sr.logger.Info("Stopping service registry")
	close(sr.stopChan)

	if sr.subscriber != nil {
		if err := sr.subscriber.Close(); err != nil {
			sr.logger.WithError(err).Warn("Error closing capability subscriber")
		}
	}
}

// listenForAnnouncements processes incoming capability announcements
func (sr *ServiceRegistry) listenForAnnouncements(ctx context.Context) {
	if sr.subscriber == nil {
		sr.logger.Error("Subscriber not initialized")
		return
	}

	ch := sr.subscriber.Channel()

	sr.logger.Info("Listening for capability announcements")

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}

			sr.logger.WithField("payload_size", len(msg.Payload)).Debug("Received capability announcement")

			var capability ServiceCapability
			if err := json.Unmarshal([]byte(msg.Payload), &capability); err != nil {
				sr.logger.WithError(err).Error("Failed to unmarshal capability announcement")
				continue
			}

			sr.updateCapability(&capability)

		case <-sr.stopChan:
			sr.logger.Info("Capability announcement listener stopped")
			return
		case <-ctx.Done():
			sr.logger.Info("Capability announcement listener cancelled")
			return
		}
	}
}

// updateCapability updates the capability information for a service
func (sr *ServiceRegistry) updateCapability(capability *ServiceCapability) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	capability.LastUpdated = time.Now()
	sr.capabilities[capability.Component] = capability
	sr.lastSeen[capability.Component] = capability.LastUpdated

	sr.logger.WithFields(logrus.Fields{
		"component":     capability.Component,
		"trigger":       capability.Trigger,
		"operations":    len(capability.Capabilities.Operations),
		"last_updated":  capability.LastUpdated,
	}).Info("Updated service capability")
}

// cleanupStaleServices removes services that haven't announced capabilities recently
func (sr *ServiceRegistry) cleanupStaleServices(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sr.removeStaleServices()
		case <-sr.stopChan:
			sr.logger.Debug("Stale service cleanup stopped")
			return
		case <-ctx.Done():
			sr.logger.Debug("Stale service cleanup cancelled")
			return
		}
	}
}

// removeStaleServices removes services that haven't been seen recently
func (sr *ServiceRegistry) removeStaleServices() {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	now := time.Now()
	removedCount := 0

	for component, lastSeen := range sr.lastSeen {
		if now.Sub(lastSeen) > sr.staleThreshold {
			delete(sr.capabilities, component)
			delete(sr.lastSeen, component)
			removedCount++

			sr.logger.WithFields(logrus.Fields{
				"component": component,
				"last_seen": lastSeen,
				"stale_age": now.Sub(lastSeen),
			}).Info("Removed stale service")
		}
	}

	if removedCount > 0 {
		sr.logger.WithField("removed_count", removedCount).Info("Cleanup completed")
	}
}

// GetServiceCapability returns the capability information for a specific service
func (sr *ServiceRegistry) GetServiceCapability(component string) (*ServiceCapability, bool) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	capability, exists := sr.capabilities[component]
	if !exists {
		return nil, false
	}

	// Check if service is still active
	if time.Since(sr.lastSeen[component]) > sr.staleThreshold {
		return nil, false
	}

	return capability, true
}

// GetAllActiveServices returns all currently active services
func (sr *ServiceRegistry) GetAllActiveServices() map[string]*ServiceCapability {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	result := make(map[string]*ServiceCapability)
	now := time.Now()

	for component, capability := range sr.capabilities {
		if lastSeen, exists := sr.lastSeen[component]; exists {
			if now.Sub(lastSeen) <= sr.staleThreshold {
				result[component] = capability
			}
		}
	}

	return result
}

// GetServicesByType returns services that have specific operation types
func (sr *ServiceRegistry) GetServicesByType(operationType string) []*ServiceCapability {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	var result []*ServiceCapability
	now := time.Now()

	for component, capability := range sr.capabilities {
		// Check if service is active
		if lastSeen, exists := sr.lastSeen[component]; exists {
			if now.Sub(lastSeen) > sr.staleThreshold {
				continue
			}
		} else {
			continue
		}

		// Check if service has the requested operation type
		for _, operation := range capability.Capabilities.Operations {
			if operation.Name == operationType {
				result = append(result, capability)
				break
			}
		}
	}

	return result
}

// GetAvailableOperations returns all available operations across all services
func (sr *ServiceRegistry) GetAvailableOperations() map[string][]Operation {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	result := make(map[string][]Operation)
	now := time.Now()

	for component, capability := range sr.capabilities {
		// Check if service is active
		if lastSeen, exists := sr.lastSeen[component]; exists {
			if now.Sub(lastSeen) > sr.staleThreshold {
				continue
			}
		} else {
			continue
		}

		result[component] = capability.Capabilities.Operations
	}

	return result
}

// GetOperationByName finds an operation by name across all services
func (sr *ServiceRegistry) GetOperationByName(operationName string) (*Operation, string, bool) {
	operations := sr.GetAvailableOperations()

	for component, ops := range operations {
		for _, op := range ops {
			if op.Name == operationName {
				return &op, component, true
			}
		}
	}

	return nil, "", false
}

// IsServiceAvailable checks if a specific service is currently available
func (sr *ServiceRegistry) IsServiceAvailable(component string) bool {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	lastSeen, exists := sr.lastSeen[component]
	if !exists {
		return false
	}

	return time.Since(lastSeen) <= sr.staleThreshold
}

// RequestCapabilityRefresh sends a refresh request to services
func (sr *ServiceRegistry) requestCapabilityRefresh(ctx context.Context, targetComponent string) {
	request := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"requester": "orchestrator",
	}

	if targetComponent != "" {
		request["component"] = targetComponent
	}

	data, err := json.Marshal(request)
	if err != nil {
		sr.logger.WithError(err).Error("Failed to marshal refresh request")
		return
	}

	if err := sr.redisClient.Publish(ctx, RefreshRequestChannel, data).Err(); err != nil {
		sr.logger.WithError(err).Error("Failed to publish refresh request")
		return
	}

	sr.logger.WithField("target", targetComponent).Info("Sent capability refresh request")
}

// RefreshCapabilities requests all services to re-announce their capabilities
func (sr *ServiceRegistry) RefreshCapabilities(ctx context.Context) error {
	sr.requestCapabilityRefresh(ctx, "")
	return nil
}

// RefreshServiceCapability requests a specific service to re-announce its capabilities
func (sr *ServiceRegistry) RefreshServiceCapability(ctx context.Context, component string) error {
	sr.requestCapabilityRefresh(ctx, component)
	return nil
}

// GetStats returns statistics about the service registry
func (sr *ServiceRegistry) GetStats() ServiceRegistryStats {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	now := time.Now()
	stats := ServiceRegistryStats{
		TotalServices:  len(sr.capabilities),
		ActiveServices: 0,
		StaleServices:  0,
		LastUpdate:     now,
		ServiceSummary: make(map[string]ServiceSummary),
	}

	for component, capability := range sr.capabilities {
		lastSeen := sr.lastSeen[component]
		isActive := now.Sub(lastSeen) <= sr.staleThreshold

		if isActive {
			stats.ActiveServices++
		} else {
			stats.StaleServices++
		}

		summary := ServiceSummary{
			Component:      component,
			OperationCount: len(capability.Capabilities.Operations),
			LastSeen:       lastSeen,
			IsActive:       isActive,
		}

		if capability.Capabilities != nil {
			summary.RequestChannel = capability.Capabilities.MessagePatterns.RequestChannel
			summary.ResponseChannel = capability.Capabilities.MessagePatterns.ResponseChannel
		}

		stats.ServiceSummary[component] = summary
	}

	return stats
}

// GenerateCapabilitySummary creates a human-readable summary of available capabilities
func (sr *ServiceRegistry) GenerateCapabilitySummary() string {
	operations := sr.GetAvailableOperations()
	
	if len(operations) == 0 {
		return "No services currently available."
	}

	summary := fmt.Sprintf("Available services (%d active):\n\n", len(operations))

	for component, ops := range operations {
		capability, _ := sr.GetServiceCapability(component)
		channels := ""
		if capability != nil && capability.Capabilities != nil {
			channels = fmt.Sprintf(" [%s -> %s]", 
				capability.Capabilities.MessagePatterns.RequestChannel,
				capability.Capabilities.MessagePatterns.ResponseChannel)
		}

		summary += fmt.Sprintf("## %s%s\n", component, channels)
		summary += fmt.Sprintf("Operations: %d\n", len(ops))

		for _, op := range ops {
			summary += fmt.Sprintf("- **%s**: %s\n", op.Name, op.Description)
			summary += fmt.Sprintf("  - Retry Safe: %t\n", op.RetrySafe)
			summary += fmt.Sprintf("  - Duration: %s\n", op.EstimatedDuration)
		}
		summary += "\n"
	}

	return summary
}

// HealthCheck verifies the registry is functioning properly
func (sr *ServiceRegistry) HealthCheck(ctx context.Context) error {
	// Check Redis connection
	if err := sr.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis connection failed: %w", err)
	}

	// Check if we have any active services
	stats := sr.GetStats()
	if stats.ActiveServices == 0 {
		return fmt.Errorf("no active services registered")
	}

	return nil
}