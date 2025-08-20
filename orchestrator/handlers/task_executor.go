package handlers

import (
	"context"
	"fmt"
	"orchestrator/models"

	"github.com/sirupsen/logrus"
)

// TaskExecutorImpl implements the TaskExecutor interface
type TaskExecutorImpl struct {
	messageCoordinator MessageCoordinator
	logger            *logrus.Logger
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(messageCoordinator MessageCoordinator) *TaskExecutorImpl {
	return &TaskExecutorImpl{
		messageCoordinator: messageCoordinator,
		logger:            logrus.New(),
	}
}

// ExecuteTask executes a single task based on its type
func (te *TaskExecutorImpl) ExecuteTask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error {
	taskState := execution.TaskStates[task.ID]
	
	te.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"task_id":      task.ID,
		"task_type":    task.Type,
	}).Info("Executing task")

	var err error
	
	switch task.Type {
	case "data":
		err = te.executeDataTask(ctx, task, execution)
	case "ai":
		err = te.executeAITask(ctx, task, execution)
	case "exec":
		err = te.executeExecTask(ctx, task, execution)
	case "parallel":
		err = te.executeParallelTask(ctx, task, execution)
	case "condition":
		err = te.executeConditionTask(ctx, task, execution)
	default:
		err = fmt.Errorf("unsupported task type: %s", task.Type)
	}

	// Store task output in state
	if err == nil {
		te.logger.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"task_id":      task.ID,
		}).Info("Task execution completed successfully")
	} else {
		te.logger.WithError(err).WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"task_id":      task.ID,
		}).Error("Task execution failed")
		
		// Store error in task state
		if taskState.Metadata == nil {
			taskState.Metadata = make(map[string]interface{})
		}
		taskState.Metadata["error"] = err.Error()
	}

	return err
}

// executeDataTask executes a data service task
func (te *TaskExecutorImpl) executeDataTask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error {
	request := &models.ServiceRequest{
		Service:    "data",
		Operation:  task.Parameters["operation"].(string),
		Parameters: task.Parameters,
		Timeout:    task.Timeout,
	}

	response, err := te.messageCoordinator.SendDataRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("data service request failed: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("data service error: %s", response.Error)
	}

	// Store response data in task state
	taskState := execution.TaskStates[task.ID]
	taskState.Output = response.Data
	if taskState.Metadata == nil {
		taskState.Metadata = make(map[string]interface{})
	}
	taskState.Metadata["service_response"] = response

	return nil
}

// executeAITask executes an AI service task
func (te *TaskExecutorImpl) executeAITask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error {
	request := &models.ServiceRequest{
		Service:    "ai",
		Operation:  "generate", // Default AI operation
		Parameters: task.Parameters,
		Timeout:    task.Timeout,
	}

	response, err := te.messageCoordinator.SendAIRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("AI service request failed: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("AI service error: %s", response.Error)
	}

	// Store response data in task state
	taskState := execution.TaskStates[task.ID]
	taskState.Output = response.Data
	if taskState.Metadata == nil {
		taskState.Metadata = make(map[string]interface{})
	}
	taskState.Metadata["service_response"] = response

	return nil
}

// executeExecTask executes a container execution task
func (te *TaskExecutorImpl) executeExecTask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error {
	// Build exec request with task parameters
	execParams := make(map[string]interface{})
	
	// Copy all task parameters to exec request
	for k, v := range task.Parameters {
		execParams[k] = v
	}
	
	// Add execution context
	execParams["execution_id"] = execution.ID
	execParams["task_id"] = task.ID
	
	// Add input data from previous tasks if specified
	if inputMappings, exists := task.Parameters["input_mappings"]; exists {
		if mappings, ok := inputMappings.(map[string]interface{}); ok {
			inputData := make(map[string]interface{})
			for targetKey, sourceTaskID := range mappings {
				if sourceID, ok := sourceTaskID.(string); ok {
					if sourceTask, exists := execution.TaskStates[sourceID]; exists {
						inputData[targetKey] = sourceTask.Output
					}
				}
			}
			execParams["input_data"] = inputData
		}
	}

	request := &models.ServiceRequest{
		Service:    "exec",
		Operation:  "execute",
		Parameters: execParams,
		Timeout:    task.Timeout,
	}

	response, err := te.messageCoordinator.SendExecRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("exec service request failed: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("exec service error: %s", response.Error)
	}

	// Store response data in task state
	taskState := execution.TaskStates[task.ID]
	taskState.Output = response.Data
	if taskState.Metadata == nil {
		taskState.Metadata = make(map[string]interface{})
	}
	taskState.Metadata["service_response"] = response

	return nil
}

// executeParallelTask executes multiple sub-tasks in parallel
func (te *TaskExecutorImpl) executeParallelTask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error {
	// Get sub-tasks from parameters
	subTasks, exists := task.Parameters["tasks"]
	if !exists {
		return fmt.Errorf("parallel task must define sub-tasks")
	}

	subTaskList, ok := subTasks.([]interface{})
	if !ok {
		return fmt.Errorf("parallel task sub-tasks must be an array")
	}

	// Execute sub-tasks in parallel (simplified implementation)
	results := make(map[string]interface{})
	
	for i, subTaskData := range subTaskList {
		subTaskMap, ok := subTaskData.(map[string]interface{})
		if !ok {
			continue
		}

		// Create a sub-task
		subTask := &models.Task{
			ID:         fmt.Sprintf("%s_sub_%d", task.ID, i),
			Name:       fmt.Sprintf("%s Sub-task %d", task.Name, i+1),
			Type:       subTaskMap["type"].(string),
			Parameters: subTaskMap["parameters"].(map[string]interface{}),
			Timeout:    task.Timeout,
		}

		// Execute sub-task
		if err := te.ExecuteTask(ctx, subTask, execution); err != nil {
			return fmt.Errorf("parallel sub-task %s failed: %w", subTask.ID, err)
		}

		// Collect result
		if subTaskState, exists := execution.TaskStates[subTask.ID]; exists {
			results[subTask.ID] = subTaskState.Output
		}
	}

	// Store combined results
	taskState := execution.TaskStates[task.ID]
	taskState.Output = map[string]interface{}{
		"results": results,
		"count":   len(results),
	}

	return nil
}

// executeConditionTask executes a conditional task
func (te *TaskExecutorImpl) executeConditionTask(ctx context.Context, task *models.Task, execution *models.WorkflowExecution) error {
	// Evaluate condition
	conditionResult, err := te.evaluateCondition(task.Condition, execution)
	if err != nil {
		return fmt.Errorf("condition evaluation failed: %w", err)
	}

	taskState := execution.TaskStates[task.ID]
	taskState.Output = map[string]interface{}{
		"condition_result": conditionResult,
		"condition":        task.Condition,
	}

	// Execute appropriate follow-up tasks based on condition result
	var followUpTasks []string
	if conditionResult {
		followUpTasks = task.OnSuccess
	} else {
		followUpTasks = task.OnFailure
	}

	if len(followUpTasks) > 0 {
		taskState.Output.(map[string]interface{})["follow_up_tasks"] = followUpTasks
	}

	te.logger.WithFields(logrus.Fields{
		"task_id":          task.ID,
		"condition":        task.Condition,
		"condition_result": conditionResult,
		"follow_up_tasks":  followUpTasks,
	}).Info("Condition task evaluated")

	return nil
}

// evaluateCondition evaluates a condition expression (simplified implementation)
func (te *TaskExecutorImpl) evaluateCondition(condition string, execution *models.WorkflowExecution) (bool, error) {
	// This is a simplified condition evaluator
	// In a real implementation, you would use a proper expression parser/evaluator
	
	te.logger.WithFields(logrus.Fields{
		"condition":    condition,
		"execution_id": execution.ID,
	}).Debug("Evaluating condition")

	// Handle simple variable existence checks
	if condition == "" {
		return true, nil
	}

	// Simple pattern matching for common conditions
	switch {
	case condition == "true":
		return true, nil
	case condition == "false":
		return false, nil
	default:
		// For complex conditions, implement a proper expression evaluator
		// For now, return true as default
		te.logger.WithField("condition", condition).Warn("Complex condition evaluation not implemented, defaulting to true")
		return true, nil
	}
}

// GetTaskResult retrieves the output of a completed task
func (te *TaskExecutorImpl) GetTaskResult(execution *models.WorkflowExecution, taskID string) (map[string]interface{}, error) {
	taskState, exists := execution.TaskStates[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found in execution", taskID)
	}

	if taskState.Status != models.StatusCompleted {
		return nil, fmt.Errorf("task %s is not completed (status: %s)", taskID, taskState.Status)
	}

	return taskState.Output, nil
}

// ValidateTask validates a task definition
func (te *TaskExecutorImpl) ValidateTask(task *models.Task) error {
	if task.ID == "" {
		return fmt.Errorf("task ID is required")
	}

	if task.Type == "" {
		return fmt.Errorf("task type is required")
	}

	// Type-specific validation
	switch task.Type {
	case "data":
		if _, exists := task.Parameters["operation"]; !exists {
			return fmt.Errorf("data task must specify operation")
		}
	case "ai":
		if _, exists := task.Parameters["prompt"]; !exists {
			return fmt.Errorf("AI task must specify prompt")
		}
	case "exec":
		if _, exists := task.Parameters["image"]; !exists {
			return fmt.Errorf("exec task must specify container image")
		}
	case "parallel":
		if _, exists := task.Parameters["tasks"]; !exists {
			return fmt.Errorf("parallel task must specify sub-tasks")
		}
	case "condition":
		if task.Condition == "" {
			return fmt.Errorf("condition task must specify condition")
		}
	default:
		return fmt.Errorf("unsupported task type: %s", task.Type)
	}

	return nil
}