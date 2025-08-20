package models

type ExecutionRequest struct {
	CorrelationID   string            `json:"correlation_id"`
	Container       ContainerSpec     `json:"container"`
	Input           InputSpec         `json:"input"`
	Output          OutputSpec        `json:"output"`
	Environment     map[string]string `json:"environment,omitempty"`
	Timeout         int               `json:"timeout,omitempty"` // seconds
	ServiceAccess   []string          `json:"service_access,omitempty"` // ["data", "ai"]
}

type ContainerSpec struct {
	Image      string   `json:"image"`
	Command    []string `json:"command,omitempty"`
	WorkingDir string   `json:"working_dir,omitempty"`
	Ports      map[string]string `json:"ports,omitempty"` // container_port:host_port
}

type InputSpec struct {
	GraphData    *GraphData     `json:"graph_data,omitempty"`
	MinioObjects []MinioObject  `json:"minio_objects,omitempty"`
	Files        []FileData     `json:"files,omitempty"`
	ConfigData   map[string]interface{} `json:"config_data,omitempty"`
}

type OutputSpec struct {
	ExpectedFiles   []string `json:"expected_files,omitempty"`
	MinioUpload     bool     `json:"minio_upload,omitempty"`
	GraphUpdate     bool     `json:"graph_update,omitempty"`
	ReturnLogs      bool     `json:"return_logs,omitempty"`
}

type GraphData struct {
	Nodes         []GraphNode         `json:"nodes"`
	Relationships []GraphRelationship `json:"relationships"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type GraphNode struct {
	ID         string                 `json:"id"`
	Labels     []string               `json:"labels"`
	Properties map[string]interface{} `json:"properties"`
}

type GraphRelationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	StartNode  string                 `json:"start_node"`
	EndNode    string                 `json:"end_node"`
	Properties map[string]interface{} `json:"properties"`
}

type MinioObject struct {
	ObjectName string `json:"object_name"`
	LocalPath  string `json:"local_path"`  // where to mount in container
}

type FileData struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Path    string `json:"path"`
}

const (
	ServiceData = "data"
	ServiceAI   = "ai"
)