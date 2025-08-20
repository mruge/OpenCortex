package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"orchestrator/models"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// AIWorkflowGenerator generates workflows using AI services
type AIWorkflowGenerator struct {
	messageCoordinator MessageCoordinator
	templateManager    TemplateManager
	serviceRegistry    ServiceRegistry
	logger            *logrus.Logger
}

// MessageCoordinator interface for AI communication
type MessageCoordinator interface {
	SendAIRequest(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error)
}

// TemplateManager interface for template management
type TemplateManager interface {
	GetTemplatesByCategory(category string) ([]*models.Template, error)
	GetTemplate(id string) (*models.Template, error)
}

// ServiceRegistry interface for capability management
type ServiceRegistry interface {
	GetAllActiveServices() map[string]*ServiceCapability
	GetAvailableOperations() map[string][]Operation
	IsServiceAvailable(component string) bool
	GenerateCapabilitySummary() string
}

// ServiceCapability represents a service's announced capabilities
type ServiceCapability struct {
	Component    string                 `json:"component"`
	Timestamp    string                 `json:"timestamp"`
	Trigger      string                 `json:"trigger"`
	Capabilities *ServiceCapabilities   `json:"capabilities"`
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

// NewAIWorkflowGenerator creates a new AI workflow generator
func NewAIWorkflowGenerator(messageCoordinator MessageCoordinator, templateManager TemplateManager, serviceRegistry ServiceRegistry) *AIWorkflowGenerator {
	return &AIWorkflowGenerator{
		messageCoordinator: messageCoordinator,
		templateManager:   templateManager,
		serviceRegistry:    serviceRegistry,
		logger:           logrus.New(),
	}
}

// GenerateWorkflow creates a workflow using AI based on user prompt
func (ai *AIWorkflowGenerator) GenerateWorkflow(ctx context.Context, request *models.AIGenerationRequest) (*models.WorkflowDefinition, error) {
	ai.logger.WithFields(logrus.Fields{
		"prompt":            request.Prompt,
		"domain":           request.Domain,
		"complexity":       request.Complexity,
		"required_services": request.RequiredServices,
	}).Info("Generating workflow with AI")

	// Check service availability if required services are specified
	if ai.serviceRegistry != nil && len(request.RequiredServices) > 0 {
		unavailableServices := []string{}
		for _, service := range request.RequiredServices {
			if !ai.serviceRegistry.IsServiceAvailable(service) {
				unavailableServices = append(unavailableServices, service)
			}
		}
		
		if len(unavailableServices) > 0 {
			ai.logger.WithField("unavailable_services", unavailableServices).Warn("Some required services are unavailable")
			// Continue with generation but include warning in the prompt
		}
	}

	// Build AI prompt with context
	prompt, err := ai.buildGenerationPrompt(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build generation prompt: %w", err)
	}

	// Send request to AI service
	aiRequest := &models.ServiceRequest{
		Service:    "ai",
		Operation:  "generate",
		Parameters: map[string]interface{}{
			"provider":         "anthropic",
			"prompt":           prompt,
			"system_message":   ai.getSystemMessage(),
			"response_format":  "yaml",
			"model":            "claude-3-sonnet",
			"max_tokens":       4000,
			"temperature":      0.3,
		},
		Timeout: 120,
	}

	response, err := ai.messageCoordinator.SendAIRequest(ctx, aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("AI generation failed: %s", response.Error)
	}

	// Extract generated content
	content, ok := response.Data["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid AI response format: missing content")
	}

	// Parse YAML workflow
	var workflow models.WorkflowDefinition
	if err := yaml.Unmarshal([]byte(content), &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse generated workflow YAML: %w", err)
	}

	// Validate and enhance generated workflow
	if err := ai.validateAndEnhanceWorkflow(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	ai.logger.WithFields(logrus.Fields{
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
		"task_count":    len(workflow.Tasks),
	}).Info("Successfully generated workflow")

	return &workflow, nil
}

// buildGenerationPrompt creates a comprehensive prompt for AI workflow generation
func (ai *AIWorkflowGenerator) buildGenerationPrompt(request *models.AIGenerationRequest) (string, error) {
	var promptBuilder strings.Builder

	// Add main request
	promptBuilder.WriteString(fmt.Sprintf("Generate a workflow for: %s\n\n", request.Prompt))

	// Add domain context
	if request.Domain != "" {
		promptBuilder.WriteString(fmt.Sprintf("Domain: %s\n", request.Domain))
	}

	// Add complexity requirements
	complexity := request.Complexity
	if complexity == "" {
		complexity = "medium"
	}
	promptBuilder.WriteString(fmt.Sprintf("Complexity level: %s\n", complexity))

	// Add service requirements
	if len(request.RequiredServices) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("Required services: %s\n", strings.Join(request.RequiredServices, ", ")))
	}

	// Add available services and their capabilities from registry
	if ai.serviceRegistry != nil {
		capabilitySummary := ai.serviceRegistry.GenerateCapabilitySummary()
		if capabilitySummary != "" {
			promptBuilder.WriteString("\nCurrently available services and operations:\n")
			promptBuilder.WriteString(capabilitySummary)
		} else {
			// Fallback to basic descriptions if no services are available
			promptBuilder.WriteString("\nAvailable service types (currently offline):\n")
			promptBuilder.WriteString("- data: Graph queries, vector search, data enrichment\n")
			promptBuilder.WriteString("- ai: Text generation, analysis, classification, summarization\n")
			promptBuilder.WriteString("- exec: Container execution, data processing, custom scripts\n")
		}
	} else {
		// Fallback if no service registry
		promptBuilder.WriteString("\nAvailable service types:\n")
		promptBuilder.WriteString("- data: Graph queries, vector search, data enrichment\n")
		promptBuilder.WriteString("- ai: Text generation, analysis, classification, summarization\n")
		promptBuilder.WriteString("- exec: Container execution, data processing, custom scripts\n")
	}

	// Add examples based on complexity
	switch complexity {
	case "simple":
		promptBuilder.WriteString(ai.getSimpleWorkflowExamples())
	case "medium":
		promptBuilder.WriteString(ai.getMediumWorkflowExamples())
	case "complex":
		promptBuilder.WriteString(ai.getComplexWorkflowExamples())
	}

	// Add template context if available
	if request.Domain != "" {
		templates, err := ai.templateManager.GetTemplatesByCategory(request.Domain)
		if err == nil && len(templates) > 0 {
			promptBuilder.WriteString("\nSimilar existing templates for reference:\n")
			for i, template := range templates {
				if i >= 3 { // Limit to 3 examples
					break
				}
				promptBuilder.WriteString(fmt.Sprintf("- %s: %s\n", template.Name, template.Description))
			}
		}
	}

	// Add format requirements
	promptBuilder.WriteString(ai.getFormatInstructions())

	return promptBuilder.String(), nil
}

// getSystemMessage returns the system message for AI workflow generation
func (ai *AIWorkflowGenerator) getSystemMessage() string {
	return `You are an expert workflow designer specializing in creating efficient, maintainable workflows for data processing, AI analysis, and automation tasks.

Your task is to generate a complete workflow definition in YAML format that:
1. Breaks down complex tasks into manageable steps
2. Properly defines task dependencies
3. Includes appropriate error handling and retry policies
4. Uses the available services effectively
5. Follows best practices for workflow design

Generate only valid YAML that matches the WorkflowDefinition schema. Do not include any explanations or comments outside the YAML.`
}

// getSimpleWorkflowExamples returns examples for simple workflows
func (ai *AIWorkflowGenerator) getSimpleWorkflowExamples() string {
	return `
Example simple workflow structure:
- Single service calls
- Linear task dependencies
- Basic error handling
- 1-3 tasks total
`
}

// getMediumWorkflowExamples returns examples for medium complexity workflows
func (ai *AIWorkflowGenerator) getMediumWorkflowExamples() string {
	return `
Example medium workflow structure:
- Multiple service integration
- Some parallel task execution
- Conditional task execution
- Variable passing between tasks
- 3-8 tasks total
`
}

// getComplexWorkflowExamples returns examples for complex workflows
func (ai *AIWorkflowGenerator) getComplexWorkflowExamples() string {
	return `
Example complex workflow structure:
- Multi-service orchestration
- Extensive parallel execution
- Complex conditional logic
- Loop constructs
- Advanced error handling and recovery
- 8+ tasks with deep dependencies
`
}

// getFormatInstructions returns formatting requirements for the AI
func (ai *AIWorkflowGenerator) getFormatInstructions() string {
	return `
Generate a complete WorkflowDefinition in YAML format with:

Required fields:
- id: unique identifier
- name: descriptive name
- description: clear explanation
- tasks: array of task definitions

Task definition requirements:
- id: unique task identifier
- name: descriptive task name
- type: one of [data, ai, exec, parallel, condition]
- parameters: service-specific parameters
- depends_on: array of task IDs (if any dependencies)

Optional but recommended:
- variables: workflow-level variables
- retry_policy: for critical tasks
- timeout: task-specific timeouts
- on_error: error handling strategy

Output only valid YAML that can be parsed directly.`
}

// validateAndEnhanceWorkflow validates and improves the generated workflow
func (ai *AIWorkflowGenerator) validateAndEnhanceWorkflow(workflow *models.WorkflowDefinition) error {
	// Ensure required fields are present
	if workflow.ID == "" {
		workflow.ID = fmt.Sprintf("ai-generated-%d", len(workflow.Name))
	}

	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(workflow.Tasks) == 0 {
		return fmt.Errorf("workflow must contain at least one task")
	}

	// Validate task definitions
	taskIDs := make(map[string]bool)
	for i, task := range workflow.Tasks {
		if task.ID == "" {
			workflow.Tasks[i].ID = fmt.Sprintf("task_%d", i+1)
		}

		if task.Name == "" {
			workflow.Tasks[i].Name = fmt.Sprintf("Task %d", i+1)
		}

		if task.Type == "" {
			return fmt.Errorf("task %s must have a type", task.ID)
		}

		// Validate task type
		validTypes := []string{"data", "ai", "exec", "parallel", "condition"}
		isValidType := false
		for _, validType := range validTypes {
			if task.Type == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			return fmt.Errorf("task %s has invalid type %s", task.ID, task.Type)
		}

		taskIDs[task.ID] = true
	}

	// Validate dependencies
	for _, task := range workflow.Tasks {
		for _, dep := range task.DependsOn {
			if !taskIDs[dep] {
				return fmt.Errorf("task %s depends on non-existent task %s", task.ID, dep)
			}
		}
	}

	// Add default retry policies for AI and external service tasks
	for i, task := range workflow.Tasks {
		if task.Type == "ai" || task.Type == "data" || task.Type == "exec" {
			if task.RetryPolicy == nil {
				workflow.Tasks[i].RetryPolicy = &models.RetryPolicy{
					MaxRetries:   2,
					BackoffType:  "exponential",
					InitialDelay: 1000, // 1 second
				}
			}
		}
	}

	// Set default timeout if not specified
	if workflow.Timeout == 0 {
		workflow.Timeout = 3600 // 1 hour default
	}

	return nil
}

// SuggestWorkflowImprovements analyzes a workflow and suggests improvements
func (ai *AIWorkflowGenerator) SuggestWorkflowImprovements(ctx context.Context, workflow *models.WorkflowDefinition) ([]string, error) {
	suggestions := make([]string, 0)

	// Analyze workflow structure
	taskTypeCount := make(map[string]int)
	hasRetryPolicies := 0
	hasTimeouts := 0
	maxDepth := 0

	for _, task := range workflow.Tasks {
		taskTypeCount[task.Type]++
		
		if task.RetryPolicy != nil {
			hasRetryPolicies++
		}
		
		if task.Timeout > 0 {
			hasTimeouts++
		}

		// Calculate dependency depth (simplified)
		depth := len(task.DependsOn)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Generate suggestions
	if hasRetryPolicies < len(workflow.Tasks)/2 {
		suggestions = append(suggestions, "Consider adding retry policies to critical tasks")
	}

	if hasTimeouts < len(workflow.Tasks)/2 {
		suggestions = append(suggestions, "Add timeout specifications to prevent hanging tasks")
	}

	if maxDepth > 5 {
		suggestions = append(suggestions, "Deep task dependencies detected - consider parallel execution where possible")
	}

	if taskTypeCount["exec"] > taskTypeCount["data"]+taskTypeCount["ai"] {
		suggestions = append(suggestions, "Heavy use of exec tasks detected - ensure proper resource management")
	}

	if len(workflow.Tasks) > 20 {
		suggestions = append(suggestions, "Large workflow detected - consider breaking into smaller workflows")
	}

	return suggestions, nil
}

// GenerateWorkflowFromTemplate creates a workflow by adapting an existing template
func (ai *AIWorkflowGenerator) GenerateWorkflowFromTemplate(ctx context.Context, templateID string, variables map[string]interface{}, customizations string) (*models.WorkflowDefinition, error) {
	// Get base template
	template, err := ai.templateManager.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Clone workflow from template
	workflow := template.Workflow
	workflow.ID = fmt.Sprintf("generated-from-%s-%d", templateID, len(customizations))

	// Apply variable substitutions
	if len(variables) > 0 {
		workflow.Variables = mergeVariables(workflow.Variables, variables)
	}

	// Apply AI-based customizations if requested
	if customizations != "" {
		customizedWorkflow, err := ai.customizeWorkflowWithAI(ctx, &workflow, customizations)
		if err != nil {
			ai.logger.WithError(err).Warn("Failed to apply AI customizations, using base template")
		} else {
			workflow = *customizedWorkflow
		}
	}

	return &workflow, nil
}

// customizeWorkflowWithAI uses AI to modify a workflow based on requirements
func (ai *AIWorkflowGenerator) customizeWorkflowWithAI(ctx context.Context, workflow *models.WorkflowDefinition, customizations string) (*models.WorkflowDefinition, error) {
	// Convert current workflow to YAML
	currentYAML, err := yaml.Marshal(workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current workflow: %w", err)
	}

	// Build customization prompt
	prompt := fmt.Sprintf(`Modify the following workflow based on these requirements: %s

Current workflow:
%s

Generate the modified workflow in YAML format, maintaining the same structure and ensuring all task dependencies remain valid.`, customizations, string(currentYAML))

	// Send to AI service
	aiRequest := &models.ServiceRequest{
		Service: "ai",
		Operation: "generate",
		Parameters: map[string]interface{}{
			"provider":        "anthropic",
			"prompt":          prompt,
			"system_message":  "You are a workflow optimization expert. Modify the given workflow according to requirements while maintaining validity.",
			"response_format": "yaml",
			"model":           "claude-3-sonnet",
			"max_tokens":      3000,
			"temperature":     0.2,
		},
		Timeout: 60,
	}

	response, err := ai.messageCoordinator.SendAIRequest(ctx, aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI customization request failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("AI customization failed: %s", response.Error)
	}

	// Parse modified workflow
	content, ok := response.Data["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid AI response format")
	}

	var customizedWorkflow models.WorkflowDefinition
	if err := yaml.Unmarshal([]byte(content), &customizedWorkflow); err != nil {
		return nil, fmt.Errorf("failed to parse customized workflow: %w", err)
	}

	// Validate customized workflow
	if err := ai.validateAndEnhanceWorkflow(&customizedWorkflow); err != nil {
		return nil, fmt.Errorf("customized workflow validation failed: %w", err)
	}

	return &customizedWorkflow, nil
}

// mergeVariables combines two variable maps with second taking precedence
func mergeVariables(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for k, v := range base {
		result[k] = v
	}
	
	for k, v := range override {
		result[k] = v
	}
	
	return result
}