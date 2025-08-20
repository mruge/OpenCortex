package capabilities

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// CapabilityManager handles service capability announcements
type CapabilityManager struct {
	component              string
	capabilities           *ServiceCapabilities
	redisClient           *redis.Client
	logger                *logrus.Logger
	refreshInterval       time.Duration
	lastCapabilityHash    string
	stopChan              chan struct{}
	refreshRequestSub     *redis.PubSub
}

// ServiceCapabilities represents the complete capability information for a service
type ServiceCapabilities struct {
	Operations       []Operation       `json:"operations"`
	MessagePatterns  MessagePatterns   `json:"message_patterns"`
}

// Operation represents a single operation that the service can perform
type Operation struct {
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	InputExample      interface{} `json:"input_example"`
	OutputExample     interface{} `json:"output_example"`
	RetrySafe         bool        `json:"retry_safe"`
	EstimatedDuration string      `json:"estimated_duration"`
}

// MessagePatterns defines how the service communicates via Redis
type MessagePatterns struct {
	RequestChannel    string `json:"request_channel"`
	ResponseChannel   string `json:"response_channel"`
	CorrelationField  string `json:"correlation_field"`
}

// CapabilityAnnouncement represents a capability announcement message
type CapabilityAnnouncement struct {
	Component    string               `json:"component"`
	Timestamp    string               `json:"timestamp"`
	Trigger      string               `json:"trigger"`
	Capabilities *ServiceCapabilities `json:"capabilities"`
}

// AnnouncementTrigger defines the possible triggers for capability announcements
type AnnouncementTrigger string

const (
	TriggerStartup        AnnouncementTrigger = "startup"
	TriggerPeriodicRefresh AnnouncementTrigger = "periodic_refresh"
	TriggerConfigChange   AnnouncementTrigger = "config_change"
	TriggerRefreshRequest AnnouncementTrigger = "refresh_request"
)

const (
	DefaultRefreshInterval = 5 * time.Minute
	AnnouncementChannel    = "service_capability_announcements"
	RefreshRequestChannel  = "capability_refresh_request"
)

// NewCapabilityManager creates a new capability manager
func NewCapabilityManager(component string, capabilities *ServiceCapabilities, redisClient *redis.Client, refreshInterval time.Duration) *CapabilityManager {
	if refreshInterval == 0 {
		refreshInterval = DefaultRefreshInterval
	}

	return &CapabilityManager{
		component:       component,
		capabilities:    capabilities,
		redisClient:     redisClient,
		logger:          logrus.WithField("component", "capability_manager"),
		refreshInterval: refreshInterval,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the capability management processes
func (cm *CapabilityManager) Start(ctx context.Context) error {
	cm.logger.WithField("component", cm.component).Info("Starting capability manager")

	// Announce capabilities on startup
	if err := cm.announceCapabilities(ctx, TriggerStartup); err != nil {
		cm.logger.WithError(err).Error("Failed to announce capabilities on startup")
		return err
	}

	// Start periodic refresh
	go cm.periodicRefresh(ctx)

	// Start refresh request listener
	go cm.listenForRefreshRequests(ctx)

	return nil
}

// Stop stops the capability manager
func (cm *CapabilityManager) Stop() {
	cm.logger.Info("Stopping capability manager")
	close(cm.stopChan)

	if cm.refreshRequestSub != nil {
		if err := cm.refreshRequestSub.Close(); err != nil {
			cm.logger.WithError(err).Warn("Error closing refresh request subscription")
		}
	}
}

// UpdateCapabilities updates the service capabilities and announces if changed
func (cm *CapabilityManager) UpdateCapabilities(ctx context.Context, newCapabilities *ServiceCapabilities) error {
	cm.capabilities = newCapabilities
	return cm.announceCapabilities(ctx, TriggerConfigChange)
}

// periodicRefresh periodically checks and announces capabilities if changed
func (cm *CapabilityManager) periodicRefresh(ctx context.Context) {
	ticker := time.NewTicker(cm.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cm.announceCapabilities(ctx, TriggerPeriodicRefresh); err != nil {
				cm.logger.WithError(err).Warn("Failed to announce capabilities during periodic refresh")
			}
		case <-cm.stopChan:
			cm.logger.Debug("Periodic refresh stopped")
			return
		case <-ctx.Done():
			cm.logger.Debug("Periodic refresh cancelled")
			return
		}
	}
}

// listenForRefreshRequests listens for refresh requests and responds
func (cm *CapabilityManager) listenForRefreshRequests(ctx context.Context) {
	cm.refreshRequestSub = cm.redisClient.Subscribe(ctx, RefreshRequestChannel)
	defer func() {
		if err := cm.refreshRequestSub.Close(); err != nil {
			cm.logger.WithError(err).Warn("Error closing refresh request subscription")
		}
	}()

	ch := cm.refreshRequestSub.Channel()

	cm.logger.Debug("Listening for capability refresh requests")

	for {
		select {
		case msg := <-ch:
			cm.logger.Debug("Received capability refresh request")
			
			// Check if the request is for this component or all components
			var request map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &request); err == nil {
				if targetComponent, exists := request["component"]; exists {
					if targetStr, ok := targetComponent.(string); ok && targetStr != cm.component {
						continue // Not for this component
					}
				}
			}

			if err := cm.announceCapabilities(ctx, TriggerRefreshRequest); err != nil {
				cm.logger.WithError(err).Warn("Failed to announce capabilities on refresh request")
			}

		case <-cm.stopChan:
			cm.logger.Debug("Refresh request listener stopped")
			return
		case <-ctx.Done():
			cm.logger.Debug("Refresh request listener cancelled")
			return
		}
	}
}

// announceCapabilities publishes capability information to Redis
func (cm *CapabilityManager) announceCapabilities(ctx context.Context, trigger AnnouncementTrigger) error {
	// Calculate hash of current capabilities
	currentHash, err := cm.calculateCapabilityHash()
	if err != nil {
		return fmt.Errorf("failed to calculate capability hash: %w", err)
	}

	// Only announce if capabilities have changed or if it's a forced trigger
	if trigger == TriggerPeriodicRefresh && currentHash == cm.lastCapabilityHash {
		cm.logger.Debug("Capabilities unchanged, skipping periodic announcement")
		return nil
	}

	// Create announcement
	announcement := &CapabilityAnnouncement{
		Component:    cm.component,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Trigger:      string(trigger),
		Capabilities: cm.capabilities,
	}

	// Marshal to JSON
	data, err := json.Marshal(announcement)
	if err != nil {
		return fmt.Errorf("failed to marshal capability announcement: %w", err)
	}

	// Publish to Redis
	if err := cm.redisClient.Publish(ctx, AnnouncementChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish capability announcement: %w", err)
	}

	// Update last hash
	cm.lastCapabilityHash = currentHash

	cm.logger.WithFields(logrus.Fields{
		"component": cm.component,
		"trigger":   trigger,
		"operations": len(cm.capabilities.Operations),
		"hash":      currentHash[:8], // Show first 8 chars of hash
	}).Info("Announced service capabilities")

	return nil
}

// calculateCapabilityHash calculates a hash of the current capabilities
func (cm *CapabilityManager) calculateCapabilityHash() (string, error) {
	data, err := json.Marshal(cm.capabilities)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// GetCapabilities returns the current capabilities
func (cm *CapabilityManager) GetCapabilities() *ServiceCapabilities {
	return cm.capabilities
}

// GetLastHash returns the hash of the last announced capabilities
func (cm *CapabilityManager) GetLastHash() string {
	return cm.lastCapabilityHash
}

// ForceAnnouncement forces a capability announcement regardless of changes
func (cm *CapabilityManager) ForceAnnouncement(ctx context.Context, trigger AnnouncementTrigger) error {
	originalHash := cm.lastCapabilityHash
	cm.lastCapabilityHash = "" // Force announcement by clearing hash
	err := cm.announceCapabilities(ctx, trigger)
	if err != nil {
		cm.lastCapabilityHash = originalHash // Restore original hash on error
	}
	return err
}