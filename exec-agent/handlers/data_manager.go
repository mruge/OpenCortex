package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"exec-agent/clients"
	"exec-agent/models"

	"github.com/sirupsen/logrus"
)

type DataManager struct {
	minioClient  *clients.MinioClient
	dockerClient *clients.DockerClient
}

func NewDataManager(minioClient *clients.MinioClient, dockerClient *clients.DockerClient) *DataManager {
	return &DataManager{
		minioClient:  minioClient,
		dockerClient: dockerClient,
	}
}

func (dm *DataManager) PrepareInputData(ctx context.Context, executionID string, input *models.InputSpec) (string, error) {
	workspacePath, err := dm.dockerClient.CreateWorkspace(executionID)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace: %v", err)
	}

	inputDir := filepath.Join(workspacePath, "input")

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"workspace":    workspacePath,
	}).Info("Preparing input data")

	// Handle graph data
	if input.GraphData != nil {
		if err := dm.writeGraphData(inputDir, input.GraphData); err != nil {
			return "", fmt.Errorf("failed to write graph data: %v", err)
		}
	}

	// Handle Minio objects
	for _, obj := range input.MinioObjects {
		localPath := filepath.Join(inputDir, obj.LocalPath)
		if err := dm.minioClient.DownloadFile(ctx, obj.ObjectName, localPath); err != nil {
			return "", fmt.Errorf("failed to download Minio object %s: %v", obj.ObjectName, err)
		}
	}

	// Handle direct file data
	for _, file := range input.Files {
		filePath := filepath.Join(inputDir, file.Path)
		
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return "", fmt.Errorf("failed to create directory for file %s: %v", file.Path, err)
		}

		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return "", fmt.Errorf("failed to write file %s: %v", file.Path, err)
		}
	}

	// Handle configuration data
	if input.ConfigData != nil {
		configPath := filepath.Join(workspacePath, "config", "config.json")
		configData, err := json.MarshalIndent(input.ConfigData, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal config data: %v", err)
		}

		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			return "", fmt.Errorf("failed to write config file: %v", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"execution_id":   executionID,
		"input_files":    len(input.Files),
		"minio_objects":  len(input.MinioObjects),
		"has_graph_data": input.GraphData != nil,
		"has_config":     input.ConfigData != nil,
	}).Info("Input data prepared successfully")

	return workspacePath, nil
}

func (dm *DataManager) ExtractOutputData(ctx context.Context, executionID, workspacePath string, output *models.OutputSpec) (*models.ExecutionResult, error) {
	result := &models.ExecutionResult{
		OutputFiles:  make([]models.OutputFile, 0),
		MinioObjects: make([]models.MinioOutputObject, 0),
		Metadata:     make(map[string]interface{}),
	}

	outputDir := filepath.Join(workspacePath, "output")

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"output_dir":   outputDir,
	}).Info("Extracting output data")

	// Check for graph update
	if output.GraphUpdate {
		graphUpdatePath := filepath.Join(outputDir, "graph_update.json")
		if graphData, err := dm.readGraphUpdate(graphUpdatePath); err == nil {
			result.GraphUpdate = graphData
			logrus.WithField("execution_id", executionID).Info("Graph update found")
		} else {
			logrus.WithError(err).Debug("No graph update found or failed to read")
		}
	}

	// Process expected files
	for _, expectedFile := range output.ExpectedFiles {
		filePath := filepath.Join(outputDir, expectedFile)
		if info, err := os.Stat(filePath); err == nil {
			outputFile := models.OutputFile{
				Name: expectedFile,
				Path: filePath,
				Size: info.Size(),
			}

			// Read content for small files (< 1MB)
			if info.Size() < 1024*1024 {
				if content, err := os.ReadFile(filePath); err == nil {
					outputFile.Content = string(content)
				}
			}

			result.OutputFiles = append(result.OutputFiles, outputFile)
		} else {
			logrus.WithFields(logrus.Fields{
				"execution_id": executionID,
				"file":         expectedFile,
			}).Warn("Expected output file not found")
		}
	}

	// Upload to Minio if requested
	if output.MinioUpload {
		minioObjects, err := dm.uploadOutputToMinio(ctx, executionID, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to upload output to Minio: %v", err)
		}
		result.MinioObjects = minioObjects
	}

	// Include logs if requested
	if output.ReturnLogs {
		logPath := filepath.Join(workspacePath, "execution.log")
		if logContent, err := os.ReadFile(logPath); err == nil {
			result.Logs = string(logContent)
		}
	}

	logrus.WithFields(logrus.Fields{
		"execution_id":   executionID,
		"output_files":   len(result.OutputFiles),
		"minio_objects":  len(result.MinioObjects),
		"has_graph_update": result.GraphUpdate != nil,
	}).Info("Output data extracted successfully")

	return result, nil
}

func (dm *DataManager) writeGraphData(inputDir string, graphData *models.GraphData) error {
	graphPath := filepath.Join(inputDir, "graph_data.json")
	
	data, err := json.MarshalIndent(graphData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph data: %v", err)
	}

	return os.WriteFile(graphPath, data, 0644)
}

func (dm *DataManager) readGraphUpdate(graphPath string) (*models.GraphData, error) {
	if _, err := os.Stat(graphPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("graph update file does not exist")
	}

	data, err := os.ReadFile(graphPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read graph update file: %v", err)
	}

	var graphData models.GraphData
	if err := json.Unmarshal(data, &graphData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graph data: %v", err)
	}

	return &graphData, nil
}

func (dm *DataManager) uploadOutputToMinio(ctx context.Context, executionID, outputDir string) ([]models.MinioOutputObject, error) {
	var minioObjects []models.MinioOutputObject

	// Check if output directory exists and has files
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return minioObjects, nil // No output directory, return empty list
	}

	// Upload entire output directory
	prefix := dm.minioClient.GenerateDirectoryPath(executionID, "output")
	uploadedObjects, err := dm.minioClient.UploadDirectory(ctx, outputDir, prefix)
	if err != nil {
		return nil, err
	}

	// Convert to MinioOutputObject format
	for _, objectName := range uploadedObjects {
		minioObjects = append(minioObjects, models.MinioOutputObject{
			ObjectName: objectName,
			Size:       0, // Size will be filled by Minio client if needed
		})
	}

	return minioObjects, nil
}

func (dm *DataManager) CleanupWorkspace(executionID string) error {
	return dm.dockerClient.CleanupWorkspace(executionID)
}