package handlers

import (
	"fmt"
	"io/fs"
	"orchestrator/models"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// TemplateManager handles workflow templates storage and retrieval
type TemplateManager struct {
	templatesDir string
	templates    map[string]*models.Template
	categories   map[string][]*models.Template
	mutex        sync.RWMutex
	logger       *logrus.Logger
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templatesDir string) *TemplateManager {
	return &TemplateManager{
		templatesDir: templatesDir,
		templates:    make(map[string]*models.Template),
		categories:   make(map[string][]*models.Template),
		logger:       logrus.New(),
	}
}

// LoadTemplates scans the templates directory and loads all templates
func (tm *TemplateManager) LoadTemplates() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.logger.WithField("templates_dir", tm.templatesDir).Info("Loading workflow templates")

	// Clear existing templates
	tm.templates = make(map[string]*models.Template)
	tm.categories = make(map[string][]*models.Template)

	// Walk through templates directory
	err := filepath.WalkDir(tm.templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			tm.logger.WithError(err).WithField("path", path).Warn("Error accessing template path")
			return nil // Continue walking
		}

		// Skip directories and non-YAML files
		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}

		// Load template file
		if err := tm.loadTemplateFile(path); err != nil {
			tm.logger.WithError(err).WithField("file", path).Error("Failed to load template file")
			// Continue loading other templates
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk templates directory: %w", err)
	}

	tm.logger.WithFields(logrus.Fields{
		"total_templates": len(tm.templates),
		"categories":      len(tm.categories),
	}).Info("Template loading completed")

	return nil
}

// loadTemplateFile loads a single template file
func (tm *TemplateManager) loadTemplateFile(filePath string) error {
	// Read file content
	content, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse YAML
	var template models.Template
	if err := yaml.Unmarshal(content, &template); err != nil {
		return fmt.Errorf("failed to parse template YAML: %w", err)
	}

	// Validate template
	if err := tm.validateTemplate(&template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Store template
	tm.templates[template.ID] = &template

	// Add to category index
	category := template.Category
	if category == "" {
		category = "general"
	}
	
	if tm.categories[category] == nil {
		tm.categories[category] = make([]*models.Template, 0)
	}
	tm.categories[category] = append(tm.categories[category], &template)

	tm.logger.WithFields(logrus.Fields{
		"template_id":   template.ID,
		"template_name": template.Name,
		"category":      category,
		"file":          filePath,
	}).Debug("Loaded template")

	return nil
}

// validateTemplate ensures template is valid
func (tm *TemplateManager) validateTemplate(template *models.Template) error {
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Workflow.ID == "" {
		template.Workflow.ID = template.ID + "_workflow"
	}

	if template.Workflow.Name == "" {
		template.Workflow.Name = template.Name
	}

	if len(template.Workflow.Tasks) == 0 {
		return fmt.Errorf("template workflow must contain at least one task")
	}

	// Validate workflow tasks
	taskIDs := make(map[string]bool)
	for _, task := range template.Workflow.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID is required")
		}

		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true

		if task.Type == "" {
			return fmt.Errorf("task type is required for task %s", task.ID)
		}
	}

	// Validate task dependencies
	for _, task := range template.Workflow.Tasks {
		for _, dep := range task.DependsOn {
			if !taskIDs[dep] {
				return fmt.Errorf("task %s depends on non-existent task %s", task.ID, dep)
			}
		}
	}

	// Validate template variables
	for _, variable := range template.Variables {
		if variable.Name == "" {
			return fmt.Errorf("template variable name is required")
		}
		if variable.Type == "" {
			return fmt.Errorf("template variable type is required for %s", variable.Name)
		}
	}

	return nil
}

// GetTemplate retrieves a template by ID
func (tm *TemplateManager) GetTemplate(id string) (*models.Template, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	template, exists := tm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template %s not found", id)
	}

	// Return a deep copy to prevent modification
	return tm.cloneTemplate(template), nil
}

// GetTemplatesByCategory returns all templates in a category
func (tm *TemplateManager) GetTemplatesByCategory(category string) ([]*models.Template, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	templates, exists := tm.categories[category]
	if !exists {
		return []*models.Template{}, nil
	}

	// Return clones to prevent modification
	result := make([]*models.Template, len(templates))
	for i, template := range templates {
		result[i] = tm.cloneTemplate(template)
	}

	return result, nil
}

// ListAllTemplates returns all available templates
func (tm *TemplateManager) ListAllTemplates() []*models.Template {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	result := make([]*models.Template, 0, len(tm.templates))
	for _, template := range tm.templates {
		result = append(result, tm.cloneTemplate(template))
	}

	return result
}

// ListCategories returns all available template categories
func (tm *TemplateManager) ListCategories() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	categories := make([]string, 0, len(tm.categories))
	for category := range tm.categories {
		categories = append(categories, category)
	}

	return categories
}

// SearchTemplates finds templates matching search criteria
func (tm *TemplateManager) SearchTemplates(query string) []*models.Template {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	query = strings.ToLower(query)
	result := make([]*models.Template, 0)

	for _, template := range tm.templates {
		if tm.templateMatches(template, query) {
			result = append(result, tm.cloneTemplate(template))
		}
	}

	return result
}

// templateMatches checks if a template matches search criteria
func (tm *TemplateManager) templateMatches(template *models.Template, query string) bool {
	// Check name, description, category
	if strings.Contains(strings.ToLower(template.Name), query) ||
		strings.Contains(strings.ToLower(template.Description), query) ||
		strings.Contains(strings.ToLower(template.Category), query) {
		return true
	}

	// Check task names and types
	for _, task := range template.Workflow.Tasks {
		if strings.Contains(strings.ToLower(task.Name), query) ||
			strings.Contains(strings.ToLower(task.Type), query) {
			return true
		}
	}

	return false
}

// CreateTemplate creates and stores a new template
func (tm *TemplateManager) CreateTemplate(template *models.Template) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Validate template
	if err := tm.validateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Check for ID conflicts
	if _, exists := tm.templates[template.ID]; exists {
		return fmt.Errorf("template with ID %s already exists", template.ID)
	}

	// Store template
	tm.templates[template.ID] = template

	// Add to category index
	category := template.Category
	if category == "" {
		category = "general"
	}
	
	if tm.categories[category] == nil {
		tm.categories[category] = make([]*models.Template, 0)
	}
	tm.categories[category] = append(tm.categories[category], template)

	tm.logger.WithFields(logrus.Fields{
		"template_id":   template.ID,
		"template_name": template.Name,
		"category":      category,
	}).Info("Created new template")

	return nil
}

// SaveTemplate persists a template to disk
func (tm *TemplateManager) SaveTemplate(template *models.Template) error {
	// Validate template
	if err := tm.validateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Generate file path
	filename := fmt.Sprintf("%s.yaml", template.ID)
	filePath := filepath.Join(tm.templatesDir, filename)

	// Marshal to YAML
	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template to YAML: %w", err)
	}

	// Write to file
	if err := writeFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	// Update in-memory storage
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.templates[template.ID] = template

	// Update category index
	category := template.Category
	if category == "" {
		category = "general"
	}

	// Remove from old category if it exists
	for cat, templates := range tm.categories {
		for i, t := range templates {
			if t.ID == template.ID && cat != category {
				tm.categories[cat] = append(templates[:i], templates[i+1:]...)
				break
			}
		}
	}

	// Add to new category
	if tm.categories[category] == nil {
		tm.categories[category] = make([]*models.Template, 0)
	}
	
	found := false
	for i, t := range tm.categories[category] {
		if t.ID == template.ID {
			tm.categories[category][i] = template
			found = true
			break
		}
	}
	
	if !found {
		tm.categories[category] = append(tm.categories[category], template)
	}

	tm.logger.WithFields(logrus.Fields{
		"template_id": template.ID,
		"file_path":   filePath,
	}).Info("Saved template to disk")

	return nil
}

// DeleteTemplate removes a template
func (tm *TemplateManager) DeleteTemplate(id string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	template, exists := tm.templates[id]
	if !exists {
		return fmt.Errorf("template %s not found", id)
	}

	// Remove from memory
	delete(tm.templates, id)

	// Remove from category index
	category := template.Category
	if category == "" {
		category = "general"
	}

	if templates, exists := tm.categories[category]; exists {
		for i, t := range templates {
			if t.ID == id {
				tm.categories[category] = append(templates[:i], templates[i+1:]...)
				break
			}
		}
	}

	tm.logger.WithField("template_id", id).Info("Deleted template")

	return nil
}

// cloneTemplate creates a deep copy of a template
func (tm *TemplateManager) cloneTemplate(template *models.Template) *models.Template {
	// Use JSON marshaling for deep copy (simple approach)
	data, err := yaml.Marshal(template)
	if err != nil {
		tm.logger.WithError(err).Error("Failed to marshal template for cloning")
		return template // Return original on error
	}

	var clone models.Template
	if err := yaml.Unmarshal(data, &clone); err != nil {
		tm.logger.WithError(err).Error("Failed to unmarshal template clone")
		return template // Return original on error
	}

	return &clone
}

// GetStats returns statistics about loaded templates
func (tm *TemplateManager) GetStats() map[string]interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	categoryStats := make(map[string]int)
	for category, templates := range tm.categories {
		categoryStats[category] = len(templates)
	}

	return map[string]interface{}{
		"total_templates":   len(tm.templates),
		"categories":        len(tm.categories),
		"category_breakdown": categoryStats,
		"templates_dir":     tm.templatesDir,
	}
}

// ReloadTemplates reloads all templates from disk
func (tm *TemplateManager) ReloadTemplates() error {
	tm.logger.Info("Reloading templates from disk")
	return tm.LoadTemplates()
}

// Utility functions for file operations  
func readFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func writeFile(filename string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(filename, data, perm)
}