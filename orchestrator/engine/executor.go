package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"orchestrator/models"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TaskExecutor handles the execution of individual tasks
type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error
}

// WorkflowExecutor orchestrates the execution of entire workflows
type WorkflowExecutor struct {
	taskExecutor    TaskExecutor
	stateManager    StateManager
	messageCoord    MessageCoordinator
	maxConcurrent   int
	logger          *logrus.Logger
}

// StateManager interface for workflow state persistence
type StateManager interface {
	SaveExecution(ctx context.Context, execution *models.WorkflowExecution) error
	LoadExecution(ctx context.Context, executionID string) (*models.WorkflowExecution, error)
	DeleteExecution(ctx context.Context, executionID string) error
	ListActiveExecutions(ctx context.Context) ([]string, error)
}

// MessageCoordinator interface for service communication
type MessageCoordinator interface {
	SendDataRequest(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error)
	SendAIRequest(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error)
	SendExecRequest(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error)
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(taskExecutor TaskExecutor, stateManager StateManager, messageCoord MessageCoordinator, maxConcurrent int) *WorkflowExecutor {
	return &WorkflowExecutor{
		taskExecutor:  taskExecutor,
		stateManager:  stateManager,
		messageCoord:  messageCoord,
		maxConcurrent: maxConcurrent,
		logger:        logrus.New(),
	}
}

// ExecuteWorkflow runs a workflow to completion
func (we *WorkflowExecutor) ExecuteWorkflow(ctx context.Context, workflow *models.WorkflowDefinition, request *models.WorkflowRequest) (*models.WorkflowResponse, error) {
	startTime := time.Now()
	
	// Create execution instance
	execution := &models.WorkflowExecution{
		ID:            generateExecutionID(),
		WorkflowID:    workflow.ID,
		CorrelationID: request.CorrelationID,
		Status:        models.StatusRunning,
		Variables:     mergeVariables(workflow.Variables, request.Variables),
		TaskStates:    make(map[string]*models.TaskState),
		StartTime:     startTime,
		Metadata:      make(map[string]interface{}),
	}

	// Initialize task states
	for _, task := range workflow.Tasks {
		execution.TaskStates[task.ID] = &models.TaskState{
			ID:         task.ID,
			Status:     models.StatusPending,
			RetryCount: 0,
			Output:     make(map[string]interface{}),
			Metadata:   make(map[string]interface{}),
		}
	}

	// Save initial state
	if err := we.stateManager.SaveExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save initial execution state: %w", err)
	}

	we.logger.WithFields(logrus.Fields{
		"execution_id":   execution.ID,
		"workflow_id":    workflow.ID,
		"correlation_id": request.CorrelationID,
	}).Info("Starting workflow execution")

	// Execute workflow
	err := we.executeDAG(ctx, workflow, execution)

	// Update final state
	endTime := time.Now()
	execution.EndTime = &endTime

	if err != nil {
		execution.Status = models.StatusFailed
		execution.Error = err.Error()
	} else {
		execution.Status = models.StatusCompleted
	}

	// Save final state
	if saveErr := we.stateManager.SaveExecution(ctx, execution); saveErr != nil {
		we.logger.WithError(saveErr).Error("Failed to save final execution state")
	}

	// Build response
	response := &models.WorkflowResponse{
		CorrelationID: request.CorrelationID,
		ExecutionID:   execution.ID,
		Status:        execution.Status,
		Success:       execution.Status == models.StatusCompleted,
		Duration:      endTime.Sub(startTime),
		Timestamp:     endTime,
	}

	if err != nil {
		response.Error = err.Error()
	} else {
		// Collect results from task outputs
		results := make(map[string]interface{})
		taskResults := make(map[string]interface{})
		
		for taskID, state := range execution.TaskStates {
			if state.Output != nil && len(state.Output) > 0 {
				taskResults[taskID] = state.Output
			}
		}
		
		response.Results = results
		response.TaskResults = taskResults
	}

	return response, err
}

// executeDAG runs the workflow using DAG-based execution
func (we *WorkflowExecutor) executeDAG(ctx context.Context, workflow *models.WorkflowDefinition, execution *models.WorkflowExecution) error {
	// Build DAG
	dag, err := NewDAG(workflow.Tasks)
	if err != nil {
		return fmt.Errorf("failed to build DAG: %w", err)
	}

	// Create execution context with timeout
	timeout := time.Duration(workflow.Timeout) * time.Second
	if timeout == 0 {
		timeout = 1 * time.Hour // default timeout
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute tasks in parallel batches
	batches := dag.GetParallelBatches()
	
	for batchIndex, batch := range batches {
		we.logger.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"batch_index":  batchIndex,
			"batch_size":   len(batch),
			"tasks":        batch,
		}).Info("Executing task batch")

		if err := we.executeBatch(execCtx, batch, workflow, execution, dag); err != nil {
			return fmt.Errorf("batch %d execution failed: %w", batchIndex, err)
		}

		// Save state after each batch
		if err := we.stateManager.SaveExecution(execCtx, execution); err != nil {
			we.logger.WithError(err).Warn("Failed to save execution state after batch")
		}

		// Check for cancellation
		if execCtx.Err() != nil {
			return fmt.Errorf("workflow execution cancelled: %w", execCtx.Err())
		}
	}

	return nil
}

// executeBatch executes a batch of parallel tasks
func (we *WorkflowExecutor) executeBatch(ctx context.Context, taskIDs []string, workflow *models.WorkflowDefinition, execution *models.WorkflowExecution, dag *DAG) error {
	// Limit concurrency
	semaphore := make(chan struct{}, we.maxConcurrent)
	errChan := make(chan error, len(taskIDs))
	var wg sync.WaitGroup

	// Execute tasks in parallel
	for _, taskID := range taskIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			task, exists := dag.GetTask(id)
			if !exists {
				errChan <- fmt.Errorf("task %s not found in DAG", id)
				return
			}

			// Execute task
			if err := we.executeTask(ctx, task, execution); err != nil {
				errChan <- fmt.Errorf("task %s failed: %w", id, err)
			}
		}(taskID)
	}

	// Wait for all tasks to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch execution errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// executeTask executes a single task with retry logic
func (we *WorkflowExecutor) executeTask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error {
	taskState := execution.TaskStates[task.ID]
	
	we.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"task_id":      task.ID,
		"task_type":    task.Type,
	}).Info("Starting task execution")

	taskState.Status = models.StatusRunning
	startTime := time.Now()
	taskState.StartTime = &startTime

	// Apply variable interpolation
	interpolatedTask, err := we.interpolateVariables(task, execution.Variables)
	if err != nil {
		taskState.Status = models.StatusFailed
		taskState.Error = fmt.Sprintf("Variable interpolation failed: %v", err)
		return err
	}

	// Execute with retry logic
	var lastErr error
	maxRetries := 0
	if task.RetryPolicy != nil {
		maxRetries = task.RetryPolicy.MaxRetries
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			taskState.Status = models.StatusRetrying
			taskState.RetryCount = attempt

			// Apply backoff delay
			delay := we.calculateBackoffDelay(task.RetryPolicy, attempt)
			we.logger.WithFields(logrus.Fields{
				"task_id": task.ID,
				"attempt": attempt,
				"delay":   delay,
			}).Info("Retrying task execution")
			
			time.Sleep(delay)
		}

		// Create task-specific context with timeout
		taskTimeout := time.Duration(task.Timeout) * time.Second
		if taskTimeout == 0 {
			taskTimeout = 5 * time.Minute // default task timeout
		}
		taskCtx, cancel := context.WithTimeout(ctx, taskTimeout)

		// Execute task
		err = we.taskExecutor.ExecuteTask(taskCtx, interpolatedTask, execution)
		cancel()

		if err == nil {
			break // Success
		}

		lastErr = err
		we.logger.WithError(err).WithFields(logrus.Fields{
			"task_id": task.ID,
			"attempt": attempt + 1,
		}).Warn("Task execution attempt failed")
	}

	// Update final task state
	endTime := time.Now()
	taskState.EndTime = &endTime

	if lastErr != nil {
		taskState.Status = models.StatusFailed
		taskState.Error = lastErr.Error()
		return lastErr
	}

	taskState.Status = models.StatusCompleted
	
	we.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"task_id":      task.ID,
		"duration":     endTime.Sub(*taskState.StartTime),
		"retry_count":  taskState.RetryCount,
	}).Info("Task execution completed")

	return nil
}

// interpolateVariables replaces variable placeholders in task parameters
func (we *WorkflowExecutor) interpolateVariables(task *models.Task, variables map[string]interface{}) (*models.Task, error) {
	// Clone task to avoid modifying original
	interpolated := *task
	interpolated.Parameters = make(map[string]interface{})

	// Interpolate parameters
	for key, value := range task.Parameters {
		interpolatedValue, err := we.interpolateValue(value, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate parameter %s: %w", key, err)
		}
		interpolated.Parameters[key] = interpolatedValue
	}

	return &interpolated, nil
}

// interpolateValue recursively interpolates variables in values
func (we *WorkflowExecutor) interpolateValue(value interface{}, variables map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return we.interpolateString(v, variables)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			interpolatedVal, err := we.interpolateValue(val, variables)
			if err != nil {
				return nil, err
			}
			result[k] = interpolatedVal
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			interpolatedVal, err := we.interpolateValue(val, variables)
			if err != nil {
				return nil, err
			}
			result[i] = interpolatedVal
		}
		return result, nil
	default:
		return value, nil
	}
}

// interpolateString replaces ${variable} placeholders in strings
func (we *WorkflowExecutor) interpolateString(s string, variables map[string]interface{}) (string, error) {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name
		varName := match[2 : len(match)-1] // Remove ${ and }
		
		if value, exists := variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}
		
		// Keep placeholder if variable not found
		return match
	}), nil
}

// calculateBackoffDelay calculates delay for retry attempts
func (we *WorkflowExecutor) calculateBackoffDelay(policy *models.RetryPolicy, attempt int) time.Duration {
	if policy == nil {
		return time.Second
	}

	delay := policy.InitialDelay

	switch policy.BackoffType {
	case "exponential":
		for i := 1; i < attempt; i++ {
			delay *= 2
		}
	case "linear":
		delay = time.Duration(int64(delay) * int64(attempt))
	// "fixed" uses initial delay as-is
	}

	if policy.MaxDelay > 0 && delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}

	return delay
}

// mergeVariables combines workflow and request variables, with request taking precedence
func mergeVariables(workflowVars, requestVars map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Copy workflow variables
	for k, v := range workflowVars {
		result[k] = v
	}
	
	// Override with request variables
	for k, v := range requestVars {
		result[k] = v
	}
	
	return result
}

// generateExecutionID creates a unique execution identifier
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d_%d", time.Now().UnixNano(), time.Now().UnixMilli()%1000)
}