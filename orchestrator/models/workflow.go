package models

import (
	"time"
)

// WorkflowRequest represents an incoming request to orchestrate a workflow
type WorkflowRequest struct {
	CorrelationID    string                 `json:"correlation_id"`
	WorkflowTemplate string                 `json:"workflow_template"`
	Variables        map[string]interface{} `json:"variables,omitempty"`
	GenerateFromAI   *AIGenerationRequest   `json:"generate_from_ai,omitempty"`
	Priority         int                    `json:"priority,omitempty"`
}

// AIGenerationRequest contains parameters for AI-generated workflow creation
type AIGenerationRequest struct {
	Prompt           string   `json:"prompt"`
	Domain           string   `json:"domain,omitempty"`
	RequiredServices []string `json:"required_services,omitempty"`
	Complexity       string   `json:"complexity,omitempty"` // simple, medium, complex
	OutputFormat     string   `json:"output_format,omitempty"`
}

// WorkflowDefinition represents a complete workflow with tasks and dependencies
type WorkflowDefinition struct {
	ID          string                 `yaml:"id" json:"id"`
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Variables   map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
	Tasks       []Task                 `yaml:"tasks" json:"tasks"`
	OnError     *ErrorHandling         `yaml:"on_error,omitempty" json:"on_error,omitempty"`
	Timeout     int                    `yaml:"timeout,omitempty" json:"timeout,omitempty"` // seconds
}

// Task represents a single step in the workflow
type Task struct {
	ID           string                 `yaml:"id" json:"id"`
	Name         string                 `yaml:"name" json:"name"`
	Type         string                 `yaml:"type" json:"type"` // data, ai, exec, parallel, condition
	DependsOn    []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Parameters   map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	RetryPolicy  *RetryPolicy           `yaml:"retry_policy,omitempty" json:"retry_policy,omitempty"`
	Timeout      int                    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Condition    string                 `yaml:"condition,omitempty" json:"condition,omitempty"`
	OnSuccess    []string               `yaml:"on_success,omitempty" json:"on_success,omitempty"`
	OnFailure    []string               `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`
	Variables    map[string]string      `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// RetryPolicy defines how tasks should be retried on failure
type RetryPolicy struct {
	MaxRetries   int           `yaml:"max_retries" json:"max_retries"`
	BackoffType  string        `yaml:"backoff_type" json:"backoff_type"` // fixed, exponential, linear
	InitialDelay time.Duration `yaml:"initial_delay" json:"initial_delay"`
	MaxDelay     time.Duration `yaml:"max_delay,omitempty" json:"max_delay,omitempty"`
}

// ErrorHandling defines global workflow error handling
type ErrorHandling struct {
	Strategy   string `yaml:"strategy" json:"strategy"` // abort, continue, retry
	MaxRetries int    `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
	Notify     string `yaml:"notify,omitempty" json:"notify,omitempty"`
}

// WorkflowExecution represents the runtime state of a workflow
type WorkflowExecution struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	CorrelationID string                 `json:"correlation_id"`
	Status        ExecutionStatus        `json:"status"`
	Variables     map[string]interface{} `json:"variables"`
	TaskStates    map[string]*TaskState  `json:"task_states"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TaskState tracks the execution state of an individual task
type TaskState struct {
	ID         string                 `json:"id"`
	Status     ExecutionStatus        `json:"status"`
	StartTime  *time.Time             `json:"start_time,omitempty"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	RetryCount int                    `json:"retry_count"`
	Error      string                 `json:"error,omitempty"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionStatus represents the current state of workflow or task execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
	StatusRetrying  ExecutionStatus = "retrying"
	StatusSkipped   ExecutionStatus = "skipped"
)

// WorkflowResponse is the response sent back after workflow execution
type WorkflowResponse struct {
	CorrelationID string                 `json:"correlation_id"`
	ExecutionID   string                 `json:"execution_id"`
	Status        ExecutionStatus        `json:"status"`
	Success       bool                   `json:"success"`
	Results       map[string]interface{} `json:"results,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Duration      time.Duration          `json:"duration"`
	TaskResults   map[string]interface{} `json:"task_results,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ServiceRequest represents a request to be sent to other services
type ServiceRequest struct {
	Service       string                 `json:"service"`       // data, ai, exec
	Operation     string                 `json:"operation"`     // service-specific operation
	CorrelationID string                 `json:"correlation_id"`
	Parameters    map[string]interface{} `json:"parameters"`
	Timeout       int                    `json:"timeout,omitempty"`
}

// ServiceResponse represents a response from other services
type ServiceResponse struct {
	CorrelationID string                 `json:"correlation_id"`
	Success       bool                   `json:"success"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Service       string                 `json:"service"`
}

// Template represents a reusable workflow template
type Template struct {
	ID          string             `yaml:"id" json:"id"`
	Name        string             `yaml:"name" json:"name"`
	Description string             `yaml:"description,omitempty" json:"description,omitempty"`
	Category    string             `yaml:"category,omitempty" json:"category,omitempty"`
	Variables   []TemplateVariable `yaml:"variables,omitempty" json:"variables,omitempty"`
	Workflow    WorkflowDefinition `yaml:"workflow" json:"workflow"`
}

// TemplateVariable defines a configurable variable in a template
type TemplateVariable struct {
	Name         string      `yaml:"name" json:"name"`
	Type         string      `yaml:"type" json:"type"` // string, int, bool, array, object
	Description  string      `yaml:"description,omitempty" json:"description,omitempty"`
	Required     bool        `yaml:"required" json:"required"`
	DefaultValue interface{} `yaml:"default,omitempty" json:"default,omitempty"`
	Options      []string    `yaml:"options,omitempty" json:"options,omitempty"`
}