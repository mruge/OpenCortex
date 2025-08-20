package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"orchestrator/capabilities"
	"orchestrator/clients"
	"orchestrator/config"
	"orchestrator/engine"
	"orchestrator/handlers"
	"orchestrator/models"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type OrchestratorServer struct {
	config             *config.Config
	redisClient        *redis.Client
	stateManager       *clients.RedisStateManager
	messageCoordinator *clients.RedisMessageCoordinator
	templateManager    *handlers.TemplateManager
	aiGenerator        *handlers.AIWorkflowGenerator
	taskExecutor       *handlers.TaskExecutorImpl
	workflowExecutor   *engine.WorkflowExecutor
	recoveryManager    *handlers.RecoveryManager
	capabilityManager  *capabilities.CapabilityManager
	logger             *logrus.Logger
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Info("Starting Orchestrator service")

	// Create orchestrator server
	server, err := NewOrchestratorServer(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create orchestrator server")
	}

	// Start server
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start orchestrator server")
	}
}

func NewOrchestratorServer(cfg *config.Config, logger *logrus.Logger) (*OrchestratorServer, error) {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
	})

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create state manager
	stateManager := clients.NewRedisStateManager(
		redisClient,
		"orchestrator",
		cfg.Orchestrator.ExecutionTTL,
	)

	// Create message coordinator
	messageCoordinator := clients.NewRedisMessageCoordinator(
		redisClient,
		clients.ServiceChannelConfig{
			RequestChannel:  cfg.Services.DataService.RequestChannel,
			ResponseChannel: cfg.Services.DataService.ResponseChannel,
			Timeout:         cfg.Services.DataService.Timeout,
		},
		clients.ServiceChannelConfig{
			RequestChannel:  cfg.Services.AIService.RequestChannel,
			ResponseChannel: cfg.Services.AIService.ResponseChannel,
			Timeout:         cfg.Services.AIService.Timeout,
		},
		clients.ServiceChannelConfig{
			RequestChannel:  cfg.Services.ExecService.RequestChannel,
			ResponseChannel: cfg.Services.ExecService.ResponseChannel,
			Timeout:         cfg.Services.ExecService.Timeout,
		},
		cfg.Services.DefaultTimeout,
	)

	// Create template manager
	templateManager := handlers.NewTemplateManager(cfg.Orchestrator.TemplatesDir)
	if err := templateManager.LoadTemplates(); err != nil {
		logger.WithError(err).Warn("Failed to load templates, continuing without templates")
	}

	// Create AI generator
	aiGenerator := handlers.NewAIWorkflowGenerator(messageCoordinator, templateManager)

	// Create task executor
	taskExecutor := handlers.NewTaskExecutor(messageCoordinator)

	// Create workflow executor
	workflowExecutor := engine.NewWorkflowExecutor(
		taskExecutor,
		stateManager,
		messageCoordinator,
		cfg.Orchestrator.MaxConcurrent,
	)

	// Create recovery manager
	recoveryManager := handlers.NewRecoveryManager(
		stateManager,
		workflowExecutor,
		templateManager,
		cfg.Orchestrator.RecoveryInterval,
	)

	// Create capability manager if enabled
	var capabilityManager *capabilities.CapabilityManager
	if cfg.Capabilities.Enabled {
		capabilityManager = capabilities.NewCapabilityManager(
			"orchestrator",
			capabilities.GetOrchestratorCapabilities(),
			redisClient,
			cfg.Capabilities.RefreshInterval,
		)
	}

	return &OrchestratorServer{
		config:             cfg,
		redisClient:        redisClient,
		stateManager:       stateManager,
		messageCoordinator: messageCoordinator,
		templateManager:    templateManager,
		aiGenerator:        aiGenerator,
		taskExecutor:       taskExecutor,
		workflowExecutor:   workflowExecutor,
		recoveryManager:    recoveryManager,
		capabilityManager:  capabilityManager,
		logger:             logger,
	}, nil
}

func (s *OrchestratorServer) Start() error {
	// Start recovery manager if enabled
	if s.config.Orchestrator.RecoveryEnabled {
		s.recoveryManager.Start(context.Background())
		defer s.recoveryManager.Stop()
	}

	// Start capability manager if enabled
	if s.capabilityManager != nil {
		if err := s.capabilityManager.Start(context.Background()); err != nil {
			s.logger.WithError(err).Error("Failed to start capability manager")
		} else {
			s.logger.Info("Capability manager started successfully")
		}
	}

	// Start message bus listener
	go s.startMessageListener()

	// Start HTTP server
	router := s.setupRoutes()
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port),
		Handler:      router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		s.logger.WithField("address", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.WithError(err).Error("HTTP server failed")
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	s.logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown server gracefully")
		return err
	}

	// Stop capability manager if running
	if s.capabilityManager != nil {
		s.capabilityManager.Stop()
		s.logger.Info("Capability manager stopped")
	}

	// Close Redis connection
	if err := s.redisClient.Close(); err != nil {
		s.logger.WithError(err).Warn("Error closing Redis connection")
	}

	// Close message coordinator
	if err := s.messageCoordinator.Close(); err != nil {
		s.logger.WithError(err).Warn("Error closing message coordinator")
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

func (s *OrchestratorServer) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/workflows", s.handleExecuteWorkflow).Methods("POST")
	api.HandleFunc("/workflows/{id}", s.handleGetWorkflow).Methods("GET")
	api.HandleFunc("/workflows/{id}/status", s.handleGetWorkflowStatus).Methods("GET")
	
	// Template routes
	api.HandleFunc("/templates", s.handleListTemplates).Methods("GET")
	api.HandleFunc("/templates/{id}", s.handleGetTemplate).Methods("GET")
	api.HandleFunc("/templates/categories/{category}", s.handleGetTemplatesByCategory).Methods("GET")
	
	// AI generation routes
	api.HandleFunc("/generate", s.handleGenerateWorkflow).Methods("POST")
	api.HandleFunc("/generate/from-template/{id}", s.handleGenerateFromTemplate).Methods("POST")
	
	// Health and status routes
	router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
	router.HandleFunc("/status", s.handleStatus).Methods("GET")
	
	// Add CORS middleware
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware(s.logger))

	return router
}

func (s *OrchestratorServer) startMessageListener() {
	s.logger.Info("Starting workflow request listener")
	
	pubsub := s.redisClient.Subscribe(context.Background(), "workflow-requests")
	defer pubsub.Close()

	ch := pubsub.Channel()
	
	for msg := range ch {
		var request models.WorkflowRequest
		if err := json.Unmarshal([]byte(msg.Payload), &request); err != nil {
			s.logger.WithError(err).Error("Failed to unmarshal workflow request")
			continue
		}

		// Process workflow request asynchronously
		go s.processWorkflowRequest(context.Background(), &request)
	}
}

func (s *OrchestratorServer) processWorkflowRequest(ctx context.Context, request *models.WorkflowRequest) {
	s.logger.WithField("correlation_id", request.CorrelationID).Info("Processing workflow request")

	var workflow *models.WorkflowDefinition
	var err error

	// Determine workflow source
	if request.GenerateFromAI != nil {
		// Generate workflow using AI
		workflow, err = s.aiGenerator.GenerateWorkflow(ctx, request.GenerateFromAI)
	} else if request.WorkflowTemplate != "" {
		// Load workflow from template
		template, templateErr := s.templateManager.GetTemplate(request.WorkflowTemplate)
		if templateErr != nil {
			s.sendWorkflowResponse(request.CorrelationID, nil, templateErr)
			return
		}
		workflow = &template.Workflow
	} else {
		s.sendWorkflowResponse(request.CorrelationID, nil, fmt.Errorf("no workflow source specified"))
		return
	}

	if err != nil {
		s.sendWorkflowResponse(request.CorrelationID, nil, err)
		return
	}

	// Execute workflow
	response, err := s.workflowExecutor.ExecuteWorkflow(ctx, workflow, request)
	s.sendWorkflowResponse(request.CorrelationID, response, err)
}

func (s *OrchestratorServer) sendWorkflowResponse(correlationID string, response *models.WorkflowResponse, err error) {
	if err != nil {
		response = &models.WorkflowResponse{
			CorrelationID: correlationID,
			Success:       false,
			Error:         err.Error(),
			Status:        models.StatusFailed,
			Timestamp:     time.Now(),
		}
	}

	responseData, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		s.logger.WithError(marshalErr).Error("Failed to marshal workflow response")
		return
	}

	if publishErr := s.redisClient.Publish(context.Background(), "workflow-responses", responseData).Err(); publishErr != nil {
		s.logger.WithError(publishErr).Error("Failed to publish workflow response")
	}
}

// HTTP Handlers

func (s *OrchestratorServer) handleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	var request models.WorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Process workflow in background
	go s.processWorkflowRequest(r.Context(), &request)

	// Return immediate response
	response := map[string]interface{}{
		"message":        "Workflow execution started",
		"correlation_id": request.CorrelationID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *OrchestratorServer) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	executionID := vars["id"]

	execution, err := s.stateManager.LoadExecution(r.Context(), executionID)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

func (s *OrchestratorServer) handleGetWorkflowStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	executionID := vars["id"]

	execution, err := s.stateManager.LoadExecution(r.Context(), executionID)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	status := map[string]interface{}{
		"execution_id": execution.ID,
		"status":       execution.Status,
		"start_time":   execution.StartTime,
		"end_time":     execution.EndTime,
		"task_count":   len(execution.TaskStates),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *OrchestratorServer) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	templates := s.templateManager.ListAllTemplates()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (s *OrchestratorServer) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	template, err := s.templateManager.GetTemplate(templateID)
	if err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func (s *OrchestratorServer) handleGetTemplatesByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]

	templates, err := s.templateManager.GetTemplatesByCategory(category)
	if err != nil {
		http.Error(w, "Failed to get templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (s *OrchestratorServer) handleGenerateWorkflow(w http.ResponseWriter, r *http.Request) {
	var request models.AIGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	workflow, err := s.aiGenerator.GenerateWorkflow(r.Context(), &request)
	if err != nil {
		http.Error(w, fmt.Sprintf("Workflow generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

func (s *OrchestratorServer) handleGenerateFromTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	var request struct {
		Variables      map[string]interface{} `json:"variables"`
		Customizations string                 `json:"customizations"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	workflow, err := s.aiGenerator.GenerateWorkflowFromTemplate(r.Context(), templateID, request.Variables, request.Customizations)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template-based generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

func (s *OrchestratorServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	// Check Redis connection
	if err := s.redisClient.Ping(r.Context()).Err(); err != nil {
		health["status"] = "unhealthy"
		health["redis_error"] = err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *OrchestratorServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	activeExecutions, _ := s.stateManager.ListActiveExecutions(r.Context())
	
	status := map[string]interface{}{
		"service":            "orchestrator",
		"status":             "running",
		"active_workflows":   len(activeExecutions),
		"templates_loaded":   len(s.templateManager.ListAllTemplates()),
		"recovery_enabled":   s.config.Orchestrator.RecoveryEnabled,
		"max_concurrent":     s.config.Orchestrator.MaxConcurrent,
		"uptime":            time.Since(time.Now()).String(), // Placeholder
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Middleware

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"url":         r.URL.Path,
				"status":      ww.statusCode,
				"duration":    time.Since(start),
				"remote_addr": r.RemoteAddr,
			}).Info("HTTP request")
		})
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}