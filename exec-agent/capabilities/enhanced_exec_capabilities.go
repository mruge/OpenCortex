package capabilities

import (
	"context"
	"fmt"
)

// EnhancedExecCapabilities generates exec agent capabilities including discovered worker images
type EnhancedExecCapabilities struct {
	imageScanner *ImageScanner
}

// NewEnhancedExecCapabilities creates a new enhanced capabilities generator
func NewEnhancedExecCapabilities(imageScanner *ImageScanner) *EnhancedExecCapabilities {
	return &EnhancedExecCapabilities{
		imageScanner: imageScanner,
	}
}

// GetExecAgentCapabilitiesWithImages returns the capability definition including all known worker images
func (e *EnhancedExecCapabilities) GetExecAgentCapabilitiesWithImages() *ServiceCapabilities {
	// Start with base exec agent capabilities
	capabilities := GetExecAgentCapabilities()
	
	// Add discovered image capabilities
	if e.imageScanner != nil {
		imageCapabilities := e.imageScanner.GetAllImageCapabilities()
		
		for imageName, imageCapability := range imageCapabilities {
			// Add each operation from the image with enhanced naming and description
			for _, operation := range imageCapability.Operations {
				enhancedOp := Operation{
					Name:              fmt.Sprintf("image_%s_%s", sanitizeImageName(imageName), operation.Name),
					Description:       fmt.Sprintf("[%s] %s", imageName, operation.Description),
					InputExample:      e.generateImageInputExample(imageName, operation),
					OutputExample:     operation.OutputExample,
					RetrySafe:         operation.RetrySafe,
					EstimatedDuration: operation.EstimatedDuration,
				}
				capabilities.Operations = append(capabilities.Operations, enhancedOp)
			}
		}
	}
	
	return capabilities
}

// generateImageInputExample creates a standardized input example for image-based operations
func (e *EnhancedExecCapabilities) generateImageInputExample(imageName string, operation Operation) map[string]interface{} {
	baseInput := map[string]interface{}{
		"correlation_id": "image-exec-001",
		"container": map[string]interface{}{
			"image":       imageName,
			"command":     []string{"/app/run.sh"}, // Default command, can be overridden
			"working_dir": "/workspace",
		},
		"input": map[string]interface{}{
			"files": []map[string]interface{}{
				{
					"name":    "input.json",
					"content": `{"task": "process_data", "parameters": {}}`,
				},
			},
			"config_data": map[string]interface{}{
				"operation_type": operation.Name,
				"image_source":   imageName,
			},
		},
		"output": map[string]interface{}{
			"expected_files": []string{"output.json"},
			"return_logs":    true,
		},
		"timeout": 600,
	}
	
	// Merge with operation-specific input example if available
	if operation.InputExample != nil {
		if opInput, ok := operation.InputExample.(map[string]interface{}); ok {
			// Merge operation input with base input
			if container, exists := opInput["container"]; exists {
				if containerMap, ok := container.(map[string]interface{}); ok {
					// Update image in container config
					containerMap["image"] = imageName
					baseInput["container"] = containerMap
				}
			}
			// Add any additional fields from operation input
			for key, value := range opInput {
				if key != "container" && key != "correlation_id" {
					baseInput[key] = value
				}
			}
		}
	}
	
	return baseInput
}

// sanitizeImageName converts an image name to a valid operation name component
func sanitizeImageName(imageName string) string {
	// Remove registry prefix and tag suffix
	name := imageName
	
	// Remove registry host (everything before the last /)
	parts := []rune{}
	foundSlash := false
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			foundSlash = true
			break
		}
		parts = append([]rune{rune(name[i])}, parts...)
	}
	
	if foundSlash {
		name = string(parts)
	}
	
	// Remove tag (everything after :)
	for i := 0; i < len(name); i++ {
		if name[i] == ':' {
			name = name[:i]
			break
		}
	}
	
	// Replace non-alphanumeric characters with underscores
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result = append(result, r)
		} else {
			result = append(result, '_')
		}
	}
	
	// Remove consecutive underscores and trim
	finalResult := make([]rune, 0, len(result))
	lastWasUnderscore := false
	for _, r := range result {
		if r == '_' {
			if !lastWasUnderscore {
				finalResult = append(finalResult, r)
			}
			lastWasUnderscore = true
		} else {
			finalResult = append(finalResult, r)
			lastWasUnderscore = false
		}
	}
	
	// Trim leading/trailing underscores
	final := string(finalResult)
	for len(final) > 0 && final[0] == '_' {
		final = final[1:]
	}
	for len(final) > 0 && final[len(final)-1] == '_' {
		final = final[:len(final)-1]
	}
	
	if final == "" {
		final = "unknown"
	}
	
	return final
}

// RefreshImageCapabilities triggers a rescan of all images and updates capabilities
func (e *EnhancedExecCapabilities) RefreshImageCapabilities(ctx context.Context) error {
	if e.imageScanner == nil {
		return fmt.Errorf("no image scanner available")
	}
	
	return e.imageScanner.ScanAllImages(ctx)
}

// GetImageCapabilitySummary returns a summary of all discovered image capabilities
func (e *EnhancedExecCapabilities) GetImageCapabilitySummary() map[string]interface{} {
	if e.imageScanner == nil {
		return map[string]interface{}{
			"image_scanning_enabled": false,
			"images_discovered":      0,
		}
	}
	
	imageCapabilities := e.imageScanner.GetAllImageCapabilities()
	
	summary := map[string]interface{}{
		"image_scanning_enabled": true,
		"images_discovered":      len(imageCapabilities),
		"images":                 make([]map[string]interface{}, 0),
	}
	
	for imageName, capability := range imageCapabilities {
		imageInfo := map[string]interface{}{
			"name":            imageName,
			"operations":      len(capability.Operations),
			"last_scanned":    capability.LastScanned,
			"size_bytes":      capability.Metadata.Size,
			"version":         capability.Metadata.Version,
			"description":     capability.Metadata.Description,
			"operation_names": make([]string, len(capability.Operations)),
		}
		
		for i, op := range capability.Operations {
			imageInfo["operation_names"].([]string)[i] = op.Name
		}
		
		summary["images"] = append(summary["images"].([]map[string]interface{}), imageInfo)
	}
	
	return summary
}