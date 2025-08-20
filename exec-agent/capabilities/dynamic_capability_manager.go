package capabilities

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// DynamicCapabilityManager extends the regular capability manager to handle dynamic image capabilities
type DynamicCapabilityManager struct {
	*CapabilityManager
	enhancedCapabilities *EnhancedExecCapabilities
	logger               *logrus.Logger
	imageRefreshInterval time.Duration
	stopImageRefresh     chan struct{}
}

// NewDynamicCapabilityManager creates a new dynamic capability manager
func NewDynamicCapabilityManager(
	component string,
	enhancedCapabilities *EnhancedExecCapabilities,
	redisClient *redis.Client,
	refreshInterval time.Duration,
	imageRefreshInterval time.Duration,
) *DynamicCapabilityManager {
	// Get initial capabilities
	initialCapabilities := enhancedCapabilities.GetExecAgentCapabilitiesWithImages()
	
	// Create the base capability manager
	baseManager := NewCapabilityManager(
		component,
		initialCapabilities,
		redisClient,
		refreshInterval,
	)
	
	return &DynamicCapabilityManager{
		CapabilityManager:    baseManager,
		enhancedCapabilities: enhancedCapabilities,
		logger:               logrus.WithField("component", "dynamic_capability_manager"),
		imageRefreshInterval: imageRefreshInterval,
		stopImageRefresh:     make(chan struct{}),
	}
}

// Start begins the dynamic capability management processes
func (dcm *DynamicCapabilityManager) Start(ctx context.Context) error {
	// Start the base capability manager
	if err := dcm.CapabilityManager.Start(ctx); err != nil {
		return err
	}
	
	// Start the image refresh monitoring
	go dcm.monitorImageChanges(ctx)
	
	dcm.logger.Info("Dynamic capability manager started")
	return nil
}

// Stop stops the dynamic capability manager
func (dcm *DynamicCapabilityManager) Stop() {
	dcm.logger.Info("Stopping dynamic capability manager")
	
	// Stop image refresh monitoring
	close(dcm.stopImageRefresh)
	
	// Stop the base capability manager
	dcm.CapabilityManager.Stop()
}

// monitorImageChanges monitors for image capability changes and updates announcements
func (dcm *DynamicCapabilityManager) monitorImageChanges(ctx context.Context) {
	if dcm.imageRefreshInterval <= 0 {
		dcm.logger.Info("Image refresh monitoring disabled")
		return
	}
	
	ticker := time.NewTicker(dcm.imageRefreshInterval)
	defer ticker.Stop()
	
	dcm.logger.WithField("interval", dcm.imageRefreshInterval).Info("Started image capability monitoring")
	
	for {
		select {
		case <-ticker.C:
			if err := dcm.refreshImageCapabilities(ctx); err != nil {
				dcm.logger.WithError(err).Warn("Failed to refresh image capabilities")
			}
			
		case <-dcm.stopImageRefresh:
			dcm.logger.Debug("Image capability monitoring stopped")
			return
			
		case <-ctx.Done():
			dcm.logger.Debug("Image capability monitoring cancelled")
			return
		}
	}
}

// refreshImageCapabilities refreshes image scanning and updates capabilities if needed
func (dcm *DynamicCapabilityManager) refreshImageCapabilities(ctx context.Context) error {
	dcm.logger.Debug("Refreshing image capabilities")
	
	// Get current capability hash before refreshing
	oldHash := dcm.CapabilityManager.GetLastHash()
	
	// Refresh image capabilities
	if err := dcm.enhancedCapabilities.RefreshImageCapabilities(ctx); err != nil {
		return err
	}
	
	// Get new capabilities
	newCapabilities := dcm.enhancedCapabilities.GetExecAgentCapabilitiesWithImages()
	
	// Update the capability manager's capabilities
	oldCapabilities := dcm.CapabilityManager.capabilities
	dcm.CapabilityManager.capabilities = newCapabilities
	
	// Calculate new hash
	newHash, err := dcm.CapabilityManager.calculateCapabilityHash()
	if err != nil {
		// Restore old capabilities on error
		dcm.CapabilityManager.capabilities = oldCapabilities
		return err
	}
	
	// If capabilities changed, announce them
	if newHash != oldHash {
		dcm.logger.WithFields(logrus.Fields{
			"old_hash":       oldHash[:8],
			"new_hash":       newHash[:8],
			"old_operations": len(oldCapabilities.Operations),
			"new_operations": len(newCapabilities.Operations),
		}).Info("Image capabilities changed, announcing update")
		
		if err := dcm.CapabilityManager.announceCapabilities(ctx, TriggerConfigChange); err != nil {
			dcm.logger.WithError(err).Error("Failed to announce updated capabilities")
			return err
		}
	} else {
		dcm.logger.Debug("Image capabilities unchanged")
	}
	
	return nil
}

// GetImageCapabilitySummary returns a summary of current image capabilities
func (dcm *DynamicCapabilityManager) GetImageCapabilitySummary() map[string]interface{} {
	return dcm.enhancedCapabilities.GetImageCapabilitySummary()
}

// ForceImageRefresh forces an immediate refresh of image capabilities
func (dcm *DynamicCapabilityManager) ForceImageRefresh(ctx context.Context) error {
	dcm.logger.Info("Forcing image capability refresh")
	return dcm.refreshImageCapabilities(ctx)
}