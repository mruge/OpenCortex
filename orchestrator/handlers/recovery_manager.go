package handlers

import (
	"context"
	"fmt"
	"orchestrator/models"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RecoveryManager handles workflow recovery and resilience
type RecoveryManager struct {
	stateManager       StateManager
	workflowExecutor   WorkflowExecutor
	templateManager    TemplateManager
	recoveryInterval   time.Duration
	logger            *logrus.Logger
	stopChan          chan bool
}

// StateManager interface for recovery operations
type StateManager interface {
	SaveExecution(ctx context.Context, execution *models.WorkflowExecution) error
	LoadExecution(ctx context.Context, executionID string) (*models.WorkflowExecution, error)
	DeleteExecution(ctx context.Context, executionID string) error
	ListActiveExecutions(ctx context.Context) ([]string, error)
	SaveTaskCheckpoint(ctx context.Context, executionID, taskID string, state *models.TaskState) error
	LoadTaskCheckpoint(ctx context.Context, executionID, taskID string) (*models.TaskState, error)
	CleanupExpiredExecutions(ctx context.Context, before time.Time) (int, error)
}

// WorkflowExecutor interface for recovery operations
type WorkflowExecutor interface {
	ExecuteWorkflow(ctx context.Context, workflow *models.WorkflowDefinition, request *models.WorkflowRequest) (*models.WorkflowResponse, error)
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(stateManager StateManager, workflowExecutor WorkflowExecutor, templateManager TemplateManager, recoveryInterval time.Duration) *RecoveryManager {
	return &RecoveryManager{
		stateManager:     stateManager,
		workflowExecutor: workflowExecutor,
		templateManager:  templateManager,
		recoveryInterval: recoveryInterval,
		logger:          logrus.New(),
		stopChan:        make(chan bool),
	}
}

// Start begins the recovery manager background processes
func (rm *RecoveryManager) Start(ctx context.Context) {
	rm.logger.WithField("recovery_interval", rm.recoveryInterval).Info("Starting recovery manager")
	
	// Start recovery loop
	go rm.recoveryLoop(ctx)
	
	// Start cleanup loop
	go rm.cleanupLoop(ctx)
}

// Stop stops the recovery manager
func (rm *RecoveryManager) Stop() {
	rm.logger.Info("Stopping recovery manager")
	close(rm.stopChan)
}

// recoveryLoop periodically checks for and recovers failed executions
func (rm *RecoveryManager) recoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(rm.recoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := rm.performRecovery(ctx); err != nil {
				rm.logger.WithError(err).Error("Recovery process failed")
			}
		case <-rm.stopChan:
			rm.logger.Info("Recovery loop stopped")
			return
		case <-ctx.Done():
			rm.logger.Info("Recovery loop cancelled")
			return
		}
	}
}

// cleanupLoop periodically cleans up expired executions
func (rm *RecoveryManager) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(rm.recoveryInterval * 2) // Run cleanup less frequently
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := rm.performCleanup(ctx); err != nil {
				rm.logger.WithError(err).Error("Cleanup process failed")
			}
		case <-rm.stopChan:
			rm.logger.Info("Cleanup loop stopped")
			return
		case <-ctx.Done():
			rm.logger.Info("Cleanup loop cancelled")
			return
		}
	}
}

// performRecovery identifies and recovers failed or stuck executions
func (rm *RecoveryManager) performRecovery(ctx context.Context) error {
	rm.logger.Debug("Starting recovery process")

	// Get list of active executions
	activeExecutions, err := rm.stateManager.ListActiveExecutions(ctx)
	if err != nil {
		return err
	}

	recoveredCount := 0
	
	for _, executionID := range activeExecutions {
		if recovered, err := rm.recoverExecution(ctx, executionID); err != nil {
			rm.logger.WithError(err).WithField("execution_id", executionID).Error("Failed to recover execution")
		} else if recovered {
			recoveredCount++
		}
	}

	if recoveredCount > 0 {
		rm.logger.WithField("recovered_count", recoveredCount).Info("Recovery process completed")
	} else {
		rm.logger.Debug("Recovery process completed - no executions needed recovery")
	}

	return nil
}

// recoverExecution attempts to recover a single execution
func (rm *RecoveryManager) recoverExecution(ctx context.Context, executionID string) (bool, error) {
	// Load execution state
	execution, err := rm.stateManager.LoadExecution(ctx, executionID)
	if err != nil {
		return false, err
	}

	// Check if execution needs recovery
	needsRecovery, reason := rm.needsRecovery(execution)
	if !needsRecovery {
		return false, nil
	}

	rm.logger.WithFields(logrus.Fields{
		"execution_id": executionID,
		"reason":      reason,
		"status":      execution.Status,
	}).Info("Recovering execution")

	// Determine recovery strategy
	strategy := rm.determineRecoveryStrategy(execution, reason)
	
	switch strategy {
	case "restart":
		return rm.restartExecution(ctx, execution)
	case "resume":
		return rm.resumeExecution(ctx, execution)
	case "fail":
		return rm.failExecution(ctx, execution, reason)
	default:
		rm.logger.WithField("strategy", strategy).Warn("Unknown recovery strategy")
		return false, nil
	}
}

// needsRecovery determines if an execution needs recovery
func (rm *RecoveryManager) needsRecovery(execution *models.WorkflowExecution) (bool, string) {
	now := time.Now()
	
	// Check if execution has been running too long
	maxExecutionTime := 4 * time.Hour // Configurable timeout
	if execution.Status == models.StatusRunning && now.Sub(execution.StartTime) > maxExecutionTime {
		return true, "execution_timeout"
	}

	// Check for stuck tasks
	for taskID, taskState := range execution.TaskStates {
		if taskState.Status == models.StatusRunning && taskState.StartTime != nil {
			taskRuntime := now.Sub(*taskState.StartTime)
			maxTaskTime := 30 * time.Minute // Configurable timeout
			
			if taskRuntime > maxTaskTime {
				return true, fmt.Sprintf("task_timeout:%s", taskID)
			}
		}
	}

	// Check for orphaned executions (running but no recent updates)
	if execution.Status == models.StatusRunning {
		// Check for recent task updates
		hasRecentActivity := false
		cutoffTime := now.Add(-15 * time.Minute)
		
		for _, taskState := range execution.TaskStates {
			if taskState.StartTime != nil && taskState.StartTime.After(cutoffTime) {
				hasRecentActivity = true
				break
			}
			if taskState.EndTime != nil && taskState.EndTime.After(cutoffTime) {
				hasRecentActivity = true
				break
			}
		}
		
		if !hasRecentActivity && now.Sub(execution.StartTime) > 30*time.Minute {
			return true, "no_recent_activity"
		}
	}

	return false, ""
}

// determineRecoveryStrategy chooses the appropriate recovery strategy
func (rm *RecoveryManager) determineRecoveryStrategy(execution *models.WorkflowExecution, reason string) string {
	switch {
	case strings.HasPrefix(reason, "task_timeout:"):
		// For task timeouts, try to resume from last checkpoint
		return "resume"
	case reason == "execution_timeout":
		// For execution timeouts, mark as failed
		return "fail"
	case reason == "no_recent_activity":
		// For orphaned executions, try to resume
		return "resume"
	default:
		return "restart"
	}
}

// restartExecution restarts an execution from the beginning
func (rm *RecoveryManager) restartExecution(ctx context.Context, execution *models.WorkflowExecution) (bool, error) {
	rm.logger.WithField("execution_id", execution.ID).Info("Restarting execution")

	// Reset execution state
	execution.Status = models.StatusPending
	execution.EndTime = nil
	
	// Reset all task states
	for taskID := range execution.TaskStates {
		execution.TaskStates[taskID] = &models.TaskState{
			ID:         taskID,
			Status:     models.StatusPending,
			RetryCount: 0,
			Output:     make(map[string]interface{}),
			Metadata:   make(map[string]interface{}),
		}
	}

	// Save updated state
	if err := rm.stateManager.SaveExecution(ctx, execution); err != nil {
		return false, err
	}

	// Note: In a real implementation, you would trigger workflow execution here
	// For now, we just mark it as ready for restart
	rm.logger.WithField("execution_id", execution.ID).Info("Execution prepared for restart")
	return true, nil
}

// resumeExecution resumes an execution from its last known good state
func (rm *RecoveryManager) resumeExecution(ctx context.Context, execution *models.WorkflowExecution) (bool, error) {
	rm.logger.WithField("execution_id", execution.ID).Info("Resuming execution")

	// Reset any running tasks to pending
	for taskID, taskState := range execution.TaskStates {
		if taskState.Status == models.StatusRunning || taskState.Status == models.StatusRetrying {
			// Try to load checkpoint
			checkpoint, err := rm.stateManager.LoadTaskCheckpoint(ctx, execution.ID, taskID)
			if err == nil && checkpoint != nil {
				// Restore from checkpoint
				execution.TaskStates[taskID] = checkpoint
				rm.logger.WithField("task_id", taskID).Debug("Restored task from checkpoint")
			} else {
				// Reset to pending
				taskState.Status = models.StatusPending
				taskState.StartTime = nil
				taskState.Error = ""
			}
		}
	}

	// Update execution status
	execution.Status = models.StatusRunning

	// Save updated state
	if err := rm.stateManager.SaveExecution(ctx, execution); err != nil {
		return false, err
	}

	rm.logger.WithField("execution_id", execution.ID).Info("Execution prepared for resume")
	return true, nil
}

// failExecution marks an execution as failed due to recovery conditions
func (rm *RecoveryManager) failExecution(ctx context.Context, execution *models.WorkflowExecution, reason string) (bool, error) {
	rm.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"reason":      reason,
	}).Info("Marking execution as failed due to recovery conditions")

	// Update execution state
	execution.Status = models.StatusFailed
	now := time.Now()
	execution.EndTime = &now
	execution.Error = fmt.Sprintf("Recovery failure: %s", reason)

	// Mark any running tasks as failed
	for taskID, taskState := range execution.TaskStates {
		if taskState.Status == models.StatusRunning || taskState.Status == models.StatusRetrying {
			taskState.Status = models.StatusFailed
			taskState.EndTime = &now
			taskState.Error = fmt.Sprintf("Task failed during recovery: %s", reason)
		}
	}

	// Save updated state
	if err := rm.stateManager.SaveExecution(ctx, execution); err != nil {
		return false, err
	}

	rm.logger.WithField("execution_id", execution.ID).Info("Execution marked as failed")
	return true, nil
}

// performCleanup removes old expired execution data
func (rm *RecoveryManager) performCleanup(ctx context.Context) error {
	rm.logger.Debug("Starting cleanup process")

	// Clean up executions older than 7 days
	cutoffTime := time.Now().Add(-7 * 24 * time.Hour)
	
	cleaned, err := rm.stateManager.CleanupExpiredExecutions(ctx, cutoffTime)
	if err != nil {
		return err
	}

	if cleaned > 0 {
		rm.logger.WithField("cleaned_count", cleaned).Info("Cleanup process completed")
	} else {
		rm.logger.Debug("Cleanup process completed - no executions cleaned")
	}

	return nil
}

// CreateCheckpoint saves a checkpoint for a task
func (rm *RecoveryManager) CreateCheckpoint(ctx context.Context, executionID, taskID string, state *models.TaskState) error {
	rm.logger.WithFields(logrus.Fields{
		"execution_id": executionID,
		"task_id":      taskID,
		"status":       state.Status,
	}).Debug("Creating task checkpoint")

	return rm.stateManager.SaveTaskCheckpoint(ctx, executionID, taskID, state)
}

// GetRecoveryStats returns statistics about recovery operations
func (rm *RecoveryManager) GetRecoveryStats(ctx context.Context) (map[string]interface{}, error) {
	activeExecutions, err := rm.stateManager.ListActiveExecutions(ctx)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"active_executions":  len(activeExecutions),
		"recovery_interval":  rm.recoveryInterval.String(),
		"last_recovery_run":  time.Now().Format(time.RFC3339),
	}

	return stats, nil
}

