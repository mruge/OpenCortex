package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"orchestrator/models"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisStateManager implements state persistence using Redis
type RedisStateManager struct {
	client     *redis.Client
	keyPrefix  string
	executionTTL time.Duration
	logger     *logrus.Logger
}

// NewRedisStateManager creates a new Redis-based state manager
func NewRedisStateManager(client *redis.Client, keyPrefix string, executionTTL time.Duration) *RedisStateManager {
	return &RedisStateManager{
		client:       client,
		keyPrefix:    keyPrefix,
		executionTTL: executionTTL,
		logger:       logrus.New(),
	}
}

// SaveExecution persists workflow execution state to Redis
func (r *RedisStateManager) SaveExecution(ctx context.Context, execution *models.WorkflowExecution) error {
	key := r.executionKey(execution.ID)
	
	data, err := json.Marshal(execution)
	if err != nil {
		return fmt.Errorf("failed to marshal execution: %w", err)
	}

	// Use pipeline for atomic updates
	pipe := r.client.TxPipeline()
	
	// Save execution data
	pipe.Set(ctx, key, data, r.executionTTL)
	
	// Add to active executions set if still running
	if execution.Status == models.StatusRunning || execution.Status == models.StatusRetrying {
		pipe.SAdd(ctx, r.activeExecutionsKey(), execution.ID)
		pipe.Expire(ctx, r.activeExecutionsKey(), r.executionTTL)
	} else {
		// Remove from active set if completed/failed
		pipe.SRem(ctx, r.activeExecutionsKey(), execution.ID)
	}

	// Add to execution index with timestamp
	indexKey := r.executionIndexKey()
	pipe.ZAdd(ctx, indexKey, &redis.Z{
		Score:  float64(execution.StartTime.Unix()),
		Member: execution.ID,
	})
	pipe.Expire(ctx, indexKey, r.executionTTL*2) // Keep index longer

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save execution state: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"status":       execution.Status,
		"key":          key,
	}).Debug("Saved execution state to Redis")

	return nil
}

// LoadExecution retrieves workflow execution state from Redis
func (r *RedisStateManager) LoadExecution(ctx context.Context, executionID string) (*models.WorkflowExecution, error) {
	key := r.executionKey(executionID)
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("execution %s not found", executionID)
		}
		return nil, fmt.Errorf("failed to load execution: %w", err)
	}

	var execution models.WorkflowExecution
	if err := json.Unmarshal([]byte(data), &execution); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"execution_id": executionID,
		"status":       execution.Status,
	}).Debug("Loaded execution state from Redis")

	return &execution, nil
}

// DeleteExecution removes workflow execution state from Redis
func (r *RedisStateManager) DeleteExecution(ctx context.Context, executionID string) error {
	pipe := r.client.TxPipeline()
	
	// Remove execution data
	pipe.Del(ctx, r.executionKey(executionID))
	
	// Remove from active executions
	pipe.SRem(ctx, r.activeExecutionsKey(), executionID)
	
	// Remove from execution index
	pipe.ZRem(ctx, r.executionIndexKey(), executionID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete execution: %w", err)
	}

	r.logger.WithField("execution_id", executionID).Debug("Deleted execution state from Redis")
	return nil
}

// ListActiveExecutions returns IDs of currently running executions
func (r *RedisStateManager) ListActiveExecutions(ctx context.Context) ([]string, error) {
	executions, err := r.client.SMembers(ctx, r.activeExecutionsKey()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list active executions: %w", err)
	}

	r.logger.WithField("count", len(executions)).Debug("Listed active executions")
	return executions, nil
}

// ListExecutionsByTimeRange returns executions within a time range
func (r *RedisStateManager) ListExecutionsByTimeRange(ctx context.Context, start, end time.Time) ([]string, error) {
	executions, err := r.client.ZRangeByScore(ctx, r.executionIndexKey(), &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", start.Unix()),
		Max: fmt.Sprintf("%d", end.Unix()),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to list executions by time range: %w", err)
	}

	return executions, nil
}

// GetExecutionStats returns statistics about stored executions
func (r *RedisStateManager) GetExecutionStats(ctx context.Context) (map[string]interface{}, error) {
	pipe := r.client.Pipeline()
	
	// Count active executions
	activeCmd := pipe.SCard(ctx, r.activeExecutionsKey())
	
	// Count total executions in index
	totalCmd := pipe.ZCard(ctx, r.executionIndexKey())
	
	// Get oldest and newest execution timestamps
	oldestCmd := pipe.ZRange(ctx, r.executionIndexKey(), 0, 0)
	newestCmd := pipe.ZRevRange(ctx, r.executionIndexKey(), 0, 0)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution stats: %w", err)
	}

	stats := map[string]interface{}{
		"active_count": activeCmd.Val(),
		"total_count":  totalCmd.Val(),
	}

	// Add timestamp info if available
	if oldest := oldestCmd.Val(); len(oldest) > 0 {
		stats["oldest_execution"] = oldest[0]
	}
	if newest := newestCmd.Val(); len(newest) > 0 {
		stats["newest_execution"] = newest[0]
	}

	return stats, nil
}

// CleanupExpiredExecutions removes old execution data
func (r *RedisStateManager) CleanupExpiredExecutions(ctx context.Context, before time.Time) (int, error) {
	// Get executions to delete
	toDelete, err := r.client.ZRangeByScore(ctx, r.executionIndexKey(), &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", before.Unix()),
	}).Result()

	if err != nil {
		return 0, fmt.Errorf("failed to find expired executions: %w", err)
	}

	if len(toDelete) == 0 {
		return 0, nil
	}

	// Delete in batches
	pipe := r.client.TxPipeline()
	
	for _, executionID := range toDelete {
		// Remove execution data
		pipe.Del(ctx, r.executionKey(executionID))
		
		// Remove from active set (if still there)
		pipe.SRem(ctx, r.activeExecutionsKey(), executionID)
		
		// Remove from index
		pipe.ZRem(ctx, r.executionIndexKey(), executionID)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired executions: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"count":       len(toDelete),
		"before_time": before,
	}).Info("Cleaned up expired executions")

	return len(toDelete), nil
}

// SaveTaskCheckpoint saves intermediate task state for recovery
func (r *RedisStateManager) SaveTaskCheckpoint(ctx context.Context, executionID, taskID string, state *models.TaskState) error {
	key := r.taskCheckpointKey(executionID, taskID)
	
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal task state: %w", err)
	}

	err = r.client.Set(ctx, key, data, r.executionTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to save task checkpoint: %w", err)
	}

	return nil
}

// LoadTaskCheckpoint retrieves intermediate task state
func (r *RedisStateManager) LoadTaskCheckpoint(ctx context.Context, executionID, taskID string) (*models.TaskState, error) {
	key := r.taskCheckpointKey(executionID, taskID)
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No checkpoint found
		}
		return nil, fmt.Errorf("failed to load task checkpoint: %w", err)
	}

	var state models.TaskState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task state: %w", err)
	}

	return &state, nil
}

// Key generation methods

func (r *RedisStateManager) executionKey(executionID string) string {
	return fmt.Sprintf("%s:execution:%s", r.keyPrefix, executionID)
}

func (r *RedisStateManager) activeExecutionsKey() string {
	return fmt.Sprintf("%s:active", r.keyPrefix)
}

func (r *RedisStateManager) executionIndexKey() string {
	return fmt.Sprintf("%s:index", r.keyPrefix)
}

func (r *RedisStateManager) taskCheckpointKey(executionID, taskID string) string {
	return fmt.Sprintf("%s:checkpoint:%s:%s", r.keyPrefix, executionID, taskID)
}

// Health check
func (r *RedisStateManager) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}