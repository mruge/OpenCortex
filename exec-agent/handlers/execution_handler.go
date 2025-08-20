package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"exec-agent/clients"
	"exec-agent/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ExecutionHandler struct {
	dockerClient *clients.DockerClient
	minioClient  *clients.MinioClient
	dataManager  *DataManager
	serviceProxy *ServiceProxy
}

func NewExecutionHandler(dockerClient *clients.DockerClient, minioClient *clients.MinioClient, serviceProxy *ServiceProxy) *ExecutionHandler {
	dataManager := NewDataManager(minioClient, dockerClient)
	
	return &ExecutionHandler{
		dockerClient: dockerClient,
		minioClient:  minioClient,
		dataManager:  dataManager,
		serviceProxy: serviceProxy,
	}
}

func (eh *ExecutionHandler) HandleRequest(ctx context.Context, data []byte) []byte {
	startTime := time.Now()
	
	var req models.ExecutionRequest
	if err := json.Unmarshal(data, &req); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal execution request")
		response := models.NewErrorResponse("", "", fmt.Sprintf("Invalid request format: %v", err), time.Since(startTime))
		responseData, _ := json.Marshal(response)
		return responseData
	}

	executionID := uuid.New().String()
	
	logrus.WithFields(logrus.Fields{
		"correlation_id": req.CorrelationID,
		"execution_id":   executionID,
		"container_image": req.Container.Image,
	}).Info("Processing execution request")

	response := eh.executeContainer(ctx, &req, executionID, startTime)
	
	responseData, err := json.Marshal(response)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal execution response")
		errorResponse := models.NewErrorResponse(req.CorrelationID, executionID, "Internal error", time.Since(startTime))
		responseData, _ = json.Marshal(errorResponse)
	}

	return responseData
}

func (eh *ExecutionHandler) executeContainer(ctx context.Context, req *models.ExecutionRequest, executionID string, startTime time.Time) *models.ExecutionResponse {
	// Set default timeout if not specified
	timeout := time.Duration(300) * time.Second // 5 minutes default
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare input data and workspace
	workspacePath, err := eh.dataManager.PrepareInputData(execCtx, executionID, &req.Input)
	if err != nil {
		return models.NewErrorResponse(req.CorrelationID, executionID, fmt.Sprintf("Failed to prepare input data: %v", err), time.Since(startTime))
	}

	// Cleanup workspace on completion
	defer func() {
		if cleanupErr := eh.dataManager.CleanupWorkspace(executionID); cleanupErr != nil {
			logrus.WithError(cleanupErr).Warn("Failed to cleanup workspace")
		}
	}()

	// Set up container configuration
	containerConfig, err := eh.buildContainerConfig(req, workspacePath, executionID)
	if err != nil {
		return models.NewErrorResponse(req.CorrelationID, executionID, fmt.Sprintf("Failed to build container config: %v", err), time.Since(startTime))
	}

	// Execute container
	execResult, err := eh.dockerClient.ExecuteContainer(execCtx, *containerConfig, executionID)
	if err != nil {
		return models.NewErrorResponse(req.CorrelationID, executionID, fmt.Sprintf("Container execution failed: %v", err), time.Since(startTime))
	}

	// Extract output data
	outputResult, err := eh.dataManager.ExtractOutputData(execCtx, executionID, workspacePath, &req.Output)
	if err != nil {
		return models.NewErrorResponse(req.CorrelationID, executionID, fmt.Sprintf("Failed to extract output data: %v", err), time.Since(startTime))
	}

	// Merge execution results
	outputResult.ExitCode = execResult.ExitCode
	outputResult.Output = execResult.Output
	if execResult.Error != "" {
		outputResult.Metadata["execution_error"] = execResult.Error
	}

	success := execResult.ExitCode == 0
	if !success && outputResult.Metadata == nil {
		outputResult.Metadata = make(map[string]interface{})
	}

	return models.NewSuccessResponse(req.CorrelationID, executionID, outputResult, time.Since(startTime))
}

func (eh *ExecutionHandler) buildContainerConfig(req *models.ExecutionRequest, workspacePath, executionID string) (*clients.ContainerConfig, error) {
	// Prepare mounts
	mounts := []clients.Mount{
		{
			Source: filepath.Join(workspacePath, "input"),
			Target: "/workspace/input",
		},
		{
			Source: filepath.Join(workspacePath, "output"),
			Target: "/workspace/output",
		},
		{
			Source: filepath.Join(workspacePath, "config"),
			Target: "/workspace/config",
		},
	}

	// Build environment variables
	environment := []string{
		fmt.Sprintf("EXECUTION_ID=%s", executionID),
		"WORKSPACE_INPUT=/workspace/input",
		"WORKSPACE_OUTPUT=/workspace/output",
		"WORKSPACE_CONFIG=/workspace/config",
	}

	// Add service proxy access if requested
	if len(req.ServiceAccess) > 0 {
		// Service proxy will be accessible at host IP
		environment = append(environment, "SERVICE_PROXY_URL=http://host.docker.internal:9000")
		
		for _, service := range req.ServiceAccess {
			switch service {
			case models.ServiceData:
				environment = append(environment, "DATA_SERVICE_URL=http://host.docker.internal:9000/data")
			case models.ServiceAI:
				environment = append(environment, "AI_SERVICE_URL=http://host.docker.internal:9000/ai")
			}
		}
	}

	// Add custom environment variables
	for key, value := range req.Environment {
		environment = append(environment, fmt.Sprintf("%s=%s", key, value))
	}

	// Set working directory
	workingDir := "/workspace"
	if req.Container.WorkingDir != "" {
		workingDir = req.Container.WorkingDir
	}

	config := &clients.ContainerConfig{
		Image:       req.Container.Image,
		Command:     req.Container.Command,
		Environment: environment,
		Mounts:      mounts,
		Ports:       req.Container.Ports,
		WorkingDir:  workingDir,
	}

	return config, nil
}

func (eh *ExecutionHandler) GetExecutionStatus(executionID string) (string, error) {
	// This could be extended to track execution status in Redis or database
	return "unknown", nil
}