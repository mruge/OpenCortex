package clients

import (
	"time"
)

// Simple mount structure
type Mount struct {
	Source string
	Target string
}

type ContainerConfig struct {
	Image       string
	Command     []string
	Environment []string
	Mounts      []Mount
	Ports       map[string]string
	WorkingDir  string
}

type ExecutionResult struct {
	ExitCode int
	Output   string
	Error    string
	Duration time.Duration
}

type DockerClient = DockerShellClient

func NewDockerClient(host, apiVersion, workDir, networkName string, timeout time.Duration) (*DockerClient, error) {
	return NewDockerShellClient(workDir, networkName, timeout)
}

