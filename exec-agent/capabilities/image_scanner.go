package capabilities

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ImageCapability represents the capabilities defined in a container image
type ImageCapability struct {
	ImageName    string      `json:"image_name"`
	Operations   []Operation `json:"operations"`
	LastScanned  time.Time   `json:"last_scanned"`
	Metadata     ImageMeta   `json:"metadata"`
}

// ImageMeta contains metadata about the image
type ImageMeta struct {
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Labels      map[string]string `json:"labels"`
	Size        int64             `json:"size"`
}

// ImageScanner scans Docker images for capability information
type ImageScanner struct {
	imageCapabilities map[string]*ImageCapability
	mutex             sync.RWMutex
	logger            *logrus.Logger
	scanInterval      time.Duration
	knownImages       []string
}

// NewImageScanner creates a new image scanner
func NewImageScanner(knownImages []string, scanInterval time.Duration) *ImageScanner {
	return &ImageScanner{
		imageCapabilities: make(map[string]*ImageCapability),
		logger:            logrus.WithField("component", "image_scanner"),
		scanInterval:      scanInterval,
		knownImages:       knownImages,
	}
}

// ScanAllImages scans all known images for capabilities
func (is *ImageScanner) ScanAllImages(ctx context.Context) error {
	is.logger.Info("Starting image capability scan")
	
	for _, imageName := range is.knownImages {
		if err := is.scanImage(ctx, imageName); err != nil {
			is.logger.WithError(err).WithField("image", imageName).Error("Failed to scan image")
			continue
		}
	}
	
	is.logger.WithField("scanned_count", len(is.imageCapabilities)).Info("Image capability scan completed")
	return nil
}

// scanImage scans a single image for capability information
func (is *ImageScanner) scanImage(ctx context.Context, imageName string) error {
	is.logger.WithField("image", imageName).Info("Scanning image for capabilities")
	
	// Pull image if not available locally
	if err := is.pullImageIfNeeded(ctx, imageName); err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	
	// Get image metadata
	metadata, err := is.getImageMetadata(ctx, imageName)
	if err != nil {
		return fmt.Errorf("failed to get metadata for image %s: %w", imageName, err)
	}
	
	// Extract capabilities from labels and files
	operations, err := is.extractCapabilities(ctx, imageName, metadata.Labels)
	if err != nil {
		return fmt.Errorf("failed to extract capabilities for image %s: %w", imageName, err)
	}
	
	// Store image capability
	is.mutex.Lock()
	is.imageCapabilities[imageName] = &ImageCapability{
		ImageName:   imageName,
		Operations:  operations,
		LastScanned: time.Now(),
		Metadata:    *metadata,
	}
	is.mutex.Unlock()
	
	is.logger.WithFields(logrus.Fields{
		"image":       imageName,
		"operations":  len(operations),
	}).Info("Image capabilities scanned successfully")
	
	return nil
}

// pullImageIfNeeded pulls the image if it's not available locally
func (is *ImageScanner) pullImageIfNeeded(ctx context.Context, imageName string) error {
	// Check if image exists locally
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageName)
	if err := cmd.Run(); err == nil {
		// Image already exists
		return nil
	}
	
	is.logger.WithField("image", imageName).Info("Pulling image")
	
	// Pull the image
	cmd = exec.CommandContext(ctx, "docker", "pull", imageName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker pull failed: %w", err)
	}
	
	return nil
}

// getImageMetadata extracts metadata from the image
func (is *ImageScanner) getImageMetadata(ctx context.Context, imageName string) (*ImageMeta, error) {
	// Get image inspection data
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker inspect failed: %w", err)
	}
	
	// Parse JSON output
	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}
	
	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no inspect data found")
	}
	
	imageData := inspectData[0]
	
	// Extract metadata
	metadata := &ImageMeta{
		Labels: make(map[string]string),
	}
	
	// Extract labels
	if labelsInterface, exists := imageData["Config"].(map[string]interface{})["Labels"]; exists && labelsInterface != nil {
		if labels, ok := labelsInterface.(map[string]interface{}); ok {
			for key, value := range labels {
				if valueStr, ok := value.(string); ok {
					metadata.Labels[key] = valueStr
				}
			}
		}
	}
	
	// Extract size
	if sizeInterface, exists := imageData["Size"]; exists {
		if size, ok := sizeInterface.(float64); ok {
			metadata.Size = int64(size)
		}
	}
	
	// Extract version, description, and author from labels
	metadata.Version = metadata.Labels["version"]
	metadata.Description = metadata.Labels["description"]
	metadata.Author = metadata.Labels["author"]
	
	return metadata, nil
}

// extractCapabilities extracts capability information from image labels and content
func (is *ImageScanner) extractCapabilities(ctx context.Context, imageName string, labels map[string]string) ([]Operation, error) {
	var operations []Operation
	
	// Check for capability labels (following a standard convention)
	if capabilitiesJSON, exists := labels["exec-agent.capabilities"]; exists {
		var labelOps []Operation
		if err := json.Unmarshal([]byte(capabilitiesJSON), &labelOps); err != nil {
			is.logger.WithError(err).WithField("image", imageName).Warn("Failed to parse capabilities label")
		} else {
			operations = append(operations, labelOps...)
		}
	}
	
	// Check for capability file in the image
	fileOps, err := is.extractCapabilitiesFromFile(ctx, imageName)
	if err != nil {
		is.logger.WithError(err).WithField("image", imageName).Debug("No capability file found in image")
	} else {
		operations = append(operations, fileOps...)
	}
	
	// Infer capabilities from known image patterns
	inferredOps := is.inferCapabilitiesFromImage(imageName, labels)
	operations = append(operations, inferredOps...)
	
	return operations, nil
}

// extractCapabilitiesFromFile extracts capabilities from a file in the image
func (is *ImageScanner) extractCapabilitiesFromFile(ctx context.Context, imageName string) ([]Operation, error) {
	// Try to extract capabilities.json from the image
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", imageName, "cat", "/app/capabilities.json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("no capabilities file found")
	}
	
	var operations []Operation
	if err := json.Unmarshal(output, &operations); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities file: %w", err)
	}
	
	return operations, nil
}

// inferCapabilitiesFromImage infers capabilities based on image name and labels
func (is *ImageScanner) inferCapabilitiesFromImage(imageName string, labels map[string]string) []Operation {
	var operations []Operation
	
	// Common capability inference patterns
	inferences := map[string]Operation{
		"python": {
			Name:              "python_execution",
			Description:       "Execute Python scripts with data processing capabilities",
			RetrySafe:         true,
			EstimatedDuration: "10s-5m",
		},
		"node": {
			Name:              "node_execution",
			Description:       "Execute Node.js applications with web service capabilities",
			RetrySafe:         true,
			EstimatedDuration: "5s-2m",
		},
		"tensorflow": {
			Name:              "ml_inference",
			Description:       "Machine learning model inference using TensorFlow",
			RetrySafe:         true,
			EstimatedDuration: "30s-10m",
		},
		"pytorch": {
			Name:              "pytorch_inference",
			Description:       "Machine learning model inference using PyTorch",
			RetrySafe:         true,
			EstimatedDuration: "30s-10m",
		},
		"data-processor": {
			Name:              "data_processing",
			Description:       "Specialized data processing and analytics",
			RetrySafe:         false,
			EstimatedDuration: "1m-30m",
		},
		"etl": {
			Name:              "etl_processing",
			Description:       "Extract, Transform, Load data processing",
			RetrySafe:         false,
			EstimatedDuration: "5m-60m",
		},
	}
	
	// Check image name for patterns
	imageLower := strings.ToLower(imageName)
	for pattern, operation := range inferences {
		if strings.Contains(imageLower, pattern) {
			// Customize operation based on image
			op := operation
			op.InputExample = is.generateInputExample(imageName)
			op.OutputExample = is.generateOutputExample(imageName)
			operations = append(operations, op)
		}
	}
	
	// Check labels for framework hints
	if framework, exists := labels["framework"]; exists {
		if operation, exists := inferences[strings.ToLower(framework)]; exists {
			op := operation
			op.Name = fmt.Sprintf("%s_custom", strings.ToLower(framework))
			op.Description = fmt.Sprintf("Custom %s-based processing", framework)
			operations = append(operations, op)
		}
	}
	
	return operations
}

// generateInputExample generates a generic input example for the image
func (is *ImageScanner) generateInputExample(imageName string) map[string]interface{} {
	return map[string]interface{}{
		"correlation_id": "worker-image-001",
		"container": map[string]interface{}{
			"image":   imageName,
			"command": []string{"/app/run.sh"},
		},
		"input": map[string]interface{}{
			"files": []map[string]interface{}{
				{
					"name":    "input.json",
					"content": `{"data": "sample input"}`,
				},
			},
			"config_data": map[string]interface{}{
				"processing_mode": "standard",
			},
		},
		"output": map[string]interface{}{
			"expected_files": []string{"output.json"},
			"return_logs":    true,
		},
		"timeout": 300,
	}
}

// generateOutputExample generates a generic output example for the image
func (is *ImageScanner) generateOutputExample(imageName string) map[string]interface{} {
	return map[string]interface{}{
		"correlation_id": "worker-image-001",
		"success":        true,
		"execution_id":   "exec_worker_001",
		"result": map[string]interface{}{
			"exit_code": 0,
			"output":    "Processing completed successfully",
			"output_files": []map[string]interface{}{
				{
					"name": "output.json",
					"path": "/workspace/output/output.json",
					"size": 256,
				},
			},
		},
	}
}

// GetAllImageCapabilities returns all discovered image capabilities
func (is *ImageScanner) GetAllImageCapabilities() map[string]*ImageCapability {
	is.mutex.RLock()
	defer is.mutex.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*ImageCapability)
	for k, v := range is.imageCapabilities {
		result[k] = v
	}
	
	return result
}

// GetImageCapability returns capabilities for a specific image
func (is *ImageScanner) GetImageCapability(imageName string) (*ImageCapability, bool) {
	is.mutex.RLock()
	defer is.mutex.RUnlock()
	
	capability, exists := is.imageCapabilities[imageName]
	return capability, exists
}

// AddKnownImage adds a new image to the known images list
func (is *ImageScanner) AddKnownImage(imageName string) {
	is.knownImages = append(is.knownImages, imageName)
}

// StartPeriodicScan starts a background goroutine that periodically scans images
func (is *ImageScanner) StartPeriodicScan(ctx context.Context) {
	if is.scanInterval <= 0 {
		is.logger.Info("Periodic scanning disabled")
		return
	}
	
	go func() {
		ticker := time.NewTicker(is.scanInterval)
		defer ticker.Stop()
		
		// Initial scan
		if err := is.ScanAllImages(ctx); err != nil {
			is.logger.WithError(err).Error("Initial image scan failed")
		}
		
		for {
			select {
			case <-ticker.C:
				if err := is.ScanAllImages(ctx); err != nil {
					is.logger.WithError(err).Error("Periodic image scan failed")
				}
			case <-ctx.Done():
				is.logger.Info("Stopping periodic image scanner")
				return
			}
		}
	}()
	
	is.logger.WithField("interval", is.scanInterval).Info("Started periodic image scanning")
}