package clients

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type DockerShellClient struct {
	workDir     string
	networkName string
	timeout     time.Duration
}

func NewDockerShellClient(workDir, networkName string, timeout time.Duration) (*DockerShellClient, error) {
	// Test docker command
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("docker command not available: %v", err)
	}

	// Ensure work directory exists
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"work_dir":     workDir,
		"network_name": networkName,
		"timeout":      timeout,
	}).Info("Docker shell client initialized")

	return &DockerShellClient{
		workDir:     workDir,
		networkName: networkName,
		timeout:     timeout,
	}, nil
}

func (d *DockerShellClient) ExecuteContainer(ctx context.Context, config ContainerConfig, executionID string) (*ExecutionResult, error) {
	startTime := time.Now()

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"image":        config.Image,
		"command":      config.Command,
	}).Info("Starting container execution via shell")

	// Build docker run command
	args := []string{"run", "--rm"}

	// Add network
	if d.networkName != "" {
		args = append(args, "--network", d.networkName)
	}

	// Add mounts
	for _, mount := range config.Mounts {
		args = append(args, "-v", fmt.Sprintf("%s:%s", mount.Source, mount.Target))
	}

	// Add environment variables
	for _, env := range config.Environment {
		args = append(args, "-e", env)
	}

	// Add working directory
	if config.WorkingDir != "" {
		args = append(args, "-w", config.WorkingDir)
	}

	// Add container name
	containerName := fmt.Sprintf("exec-agent-%s", executionID)
	args = append(args, "--name", containerName)

	// Add image
	args = append(args, config.Image)

	// Add command
	args = append(args, config.Command...)

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"docker_args":  strings.Join(args, " "),
	}).Debug("Docker command prepared")

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "docker", args...)
	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)

	result := &ExecutionResult{
		Output:   string(output),
		Duration: duration,
	}

	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			result.ExitCode = -1
			result.Error = "Container execution timeout"
		} else {
			result.ExitCode = -1
			result.Error = fmt.Sprintf("Container execution failed: %v", err)
		}
	} else {
		result.ExitCode = cmd.ProcessState.ExitCode()
		if result.ExitCode != 0 {
			result.Error = fmt.Sprintf("Container exited with code %d", result.ExitCode)
		}
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"exit_code":    result.ExitCode,
		"duration":     duration,
	}).Info("Container execution completed")

	return result, nil
}

func (d *DockerShellClient) CreateWorkspace(executionID string) (string, error) {
	workspacePath := filepath.Join(d.workDir, executionID)
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace: %v", err)
	}

	// Create standard directories
	dirs := []string{"input", "output", "config"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(workspacePath, dir), 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return workspacePath, nil
}

func (d *DockerShellClient) CleanupWorkspace(executionID string) error {
	workspacePath := filepath.Join(d.workDir, executionID)
	return os.RemoveAll(workspacePath)
}

func (d *DockerShellClient) Close() error {
	return nil
}