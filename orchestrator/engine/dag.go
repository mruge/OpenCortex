package engine

import (
	"fmt"
	"orchestrator/models"
	"sort"
)

// DAG represents a Directed Acyclic Graph for task dependencies
type DAG struct {
	tasks        map[string]*models.Task
	dependencies map[string][]string
	dependents   map[string][]string
	sorted       []string
}

// NewDAG creates a new DAG from workflow tasks
func NewDAG(tasks []models.Task) (*DAG, error) {
	dag := &DAG{
		tasks:        make(map[string]*models.Task),
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
	}

	// Build task map and dependency graph
	for i := range tasks {
		task := &tasks[i]
		dag.tasks[task.ID] = task
		dag.dependencies[task.ID] = task.DependsOn

		// Build reverse dependency map (dependents)
		for _, dep := range task.DependsOn {
			if _, exists := dag.dependents[dep]; !exists {
				dag.dependents[dep] = make([]string, 0)
			}
			dag.dependents[dep] = append(dag.dependents[dep], task.ID)
		}
	}

	// Validate all dependencies exist
	if err := dag.validateDependencies(); err != nil {
		return nil, err
	}

	// Check for cycles
	if err := dag.detectCycles(); err != nil {
		return nil, err
	}

	// Perform topological sort
	if err := dag.topologicalSort(); err != nil {
		return nil, err
	}

	return dag, nil
}

// validateDependencies ensures all referenced dependencies exist as tasks
func (dag *DAG) validateDependencies() error {
	for taskID, deps := range dag.dependencies {
		for _, dep := range deps {
			if _, exists := dag.tasks[dep]; !exists {
				return fmt.Errorf("task %s depends on non-existent task %s", taskID, dep)
			}
		}
	}
	return nil
}

// detectCycles uses DFS to detect circular dependencies
func (dag *DAG) detectCycles() error {
	white := make(map[string]bool)
	gray := make(map[string]bool)
	black := make(map[string]bool)

	// Initialize all nodes as white (unvisited)
	for taskID := range dag.tasks {
		white[taskID] = true
	}

	// DFS from each white node
	for taskID := range white {
		if err := dag.dfsVisit(taskID, white, gray, black); err != nil {
			return err
		}
	}

	return nil
}

// dfsVisit performs depth-first search to detect cycles
func (dag *DAG) dfsVisit(taskID string, white, gray, black map[string]bool) error {
	// Move from white to gray
	delete(white, taskID)
	gray[taskID] = true

	// Visit all dependents
	for _, dependent := range dag.dependents[taskID] {
		if gray[dependent] {
			return fmt.Errorf("circular dependency detected: task %s depends on %s", dependent, taskID)
		}
		if white[dependent] {
			if err := dag.dfsVisit(dependent, white, gray, black); err != nil {
				return err
			}
		}
	}

	// Move from gray to black
	delete(gray, taskID)
	black[taskID] = true

	return nil
}

// topologicalSort performs topological sorting using Kahn's algorithm
func (dag *DAG) topologicalSort() error {
	inDegree := make(map[string]int)
	queue := make([]string, 0)

	// Calculate in-degree for each node
	for taskID := range dag.tasks {
		inDegree[taskID] = len(dag.dependencies[taskID])
		if inDegree[taskID] == 0 {
			queue = append(queue, taskID)
		}
	}

	// Process queue
	result := make([]string, 0, len(dag.tasks))

	for len(queue) > 0 {
		// Remove task with no dependencies
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Update in-degree of dependents
		for _, dependent := range dag.dependents[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check if all tasks were processed
	if len(result) != len(dag.tasks) {
		return fmt.Errorf("circular dependency detected during topological sort")
	}

	dag.sorted = result
	return nil
}

// GetExecutionOrder returns tasks in topologically sorted order
func (dag *DAG) GetExecutionOrder() []string {
	return dag.sorted
}

// GetReadyTasks returns tasks that have no pending dependencies
func (dag *DAG) GetReadyTasks(completedTasks map[string]bool) []string {
	ready := make([]string, 0)

	for _, taskID := range dag.sorted {
		// Skip if already completed
		if completedTasks[taskID] {
			continue
		}

		// Check if all dependencies are completed
		allDepsCompleted := true
		for _, dep := range dag.dependencies[taskID] {
			if !completedTasks[dep] {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			ready = append(ready, taskID)
		}
	}

	return ready
}

// GetTask returns a task by ID
func (dag *DAG) GetTask(taskID string) (*models.Task, bool) {
	task, exists := dag.tasks[taskID]
	return task, exists
}

// GetDependencies returns the dependencies of a task
func (dag *DAG) GetDependencies(taskID string) []string {
	return dag.dependencies[taskID]
}

// GetDependents returns the dependents of a task
func (dag *DAG) GetDependents(taskID string) []string {
	return dag.dependents[taskID]
}

// GetParallelBatches groups tasks that can run in parallel
func (dag *DAG) GetParallelBatches() [][]string {
	batches := make([][]string, 0)
	remaining := make(map[string]bool)
	
	// Initialize remaining tasks
	for taskID := range dag.tasks {
		remaining[taskID] = true
	}

	// Create batches of parallel tasks
	for len(remaining) > 0 {
		batch := make([]string, 0)
		completed := make(map[string]bool)
		
		// Add tasks that were completed in previous batches
		for taskID := range dag.tasks {
			if !remaining[taskID] {
				completed[taskID] = true
			}
		}

		// Find tasks ready to execute
		for taskID := range remaining {
			allDepsCompleted := true
			for _, dep := range dag.dependencies[taskID] {
				if !completed[dep] {
					allDepsCompleted = false
					break
				}
			}

			if allDepsCompleted {
				batch = append(batch, taskID)
			}
		}

		if len(batch) == 0 {
			// This shouldn't happen if DAG is valid
			break
		}

		// Sort batch for deterministic ordering
		sort.Strings(batch)
		batches = append(batches, batch)

		// Remove batch tasks from remaining
		for _, taskID := range batch {
			delete(remaining, taskID)
		}
	}

	return batches
}

// Clone creates a deep copy of the DAG
func (dag *DAG) Clone() *DAG {
	clone := &DAG{
		tasks:        make(map[string]*models.Task),
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
		sorted:       make([]string, len(dag.sorted)),
	}

	// Copy tasks
	for id, task := range dag.tasks {
		taskCopy := *task
		clone.tasks[id] = &taskCopy
	}

	// Copy dependencies
	for id, deps := range dag.dependencies {
		clone.dependencies[id] = make([]string, len(deps))
		copy(clone.dependencies[id], deps)
	}

	// Copy dependents
	for id, deps := range dag.dependents {
		clone.dependents[id] = make([]string, len(deps))
		copy(clone.dependents[id], deps)
	}

	// Copy sorted order
	copy(clone.sorted, dag.sorted)

	return clone
}