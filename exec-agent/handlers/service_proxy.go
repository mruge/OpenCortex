package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type ServiceProxy struct {
	dataAbstractorURL string
	aiAbstractorURL   string
	server            *http.Server
	redisPublisher    func(channel string, data interface{}) error
}

func NewServiceProxy(port int, dataAbstractorURL, aiAbstractorURL string, redisPublisher func(string, interface{}) error) *ServiceProxy {
	return &ServiceProxy{
		dataAbstractorURL: dataAbstractorURL,
		aiAbstractorURL:   aiAbstractorURL,
		redisPublisher:    redisPublisher,
	}
}

func (sp *ServiceProxy) Start(ctx context.Context, port int) error {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", sp.healthHandler).Methods("GET")

	// Data abstractor proxy endpoints
	router.PathPrefix("/data/").HandlerFunc(sp.dataProxyHandler).Methods("POST", "GET", "PUT", "DELETE")

	// AI abstractor proxy endpoints  
	router.PathPrefix("/ai/").HandlerFunc(sp.aiProxyHandler).Methods("POST", "GET", "PUT", "DELETE")

	// Direct Redis messaging endpoints
	router.HandleFunc("/data/query", sp.dataQueryHandler).Methods("POST")
	router.HandleFunc("/ai/query", sp.aiQueryHandler).Methods("POST")

	sp.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	logrus.WithFields(logrus.Fields{
		"port":                port,
		"data_abstractor_url": sp.dataAbstractorURL,
		"ai_abstractor_url":   sp.aiAbstractorURL,
	}).Info("Starting service proxy server")

	go func() {
		<-ctx.Done()
		logrus.Info("Shutting down service proxy server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		sp.server.Shutdown(shutdownCtx)
	}()

	if err := sp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("service proxy server error: %v", err)
	}

	return nil
}

func (sp *ServiceProxy) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": map[string]string{
			"data_abstractor": sp.dataAbstractorURL,
			"ai_abstractor":   sp.aiAbstractorURL,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (sp *ServiceProxy) dataProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Remove /data prefix from path
	targetPath := strings.TrimPrefix(r.URL.Path, "/data")
	targetURL := sp.dataAbstractorURL + targetPath

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	sp.proxyRequest(w, r, targetURL)
}

func (sp *ServiceProxy) aiProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Remove /ai prefix from path
	targetPath := strings.TrimPrefix(r.URL.Path, "/ai")
	targetURL := sp.aiAbstractorURL + targetPath

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	sp.proxyRequest(w, r, targetURL)
}

func (sp *ServiceProxy) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// Create new request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute request: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		logrus.WithError(err).Error("Failed to copy response body")
	}
}

func (sp *ServiceProxy) dataQueryHandler(w http.ResponseWriter, r *http.Request) {
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Add correlation ID if not present
	if _, exists := request["correlation_id"]; !exists {
		request["correlation_id"] = fmt.Sprintf("proxy-%d", time.Now().UnixNano())
	}

	// Publish to Redis
	if err := sp.redisPublisher("data-requests", request); err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish message: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"status":         "queued",
		"correlation_id": request["correlation_id"],
		"message":        "Request queued successfully. Monitor 'data-responses' channel for results.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (sp *ServiceProxy) aiQueryHandler(w http.ResponseWriter, r *http.Request) {
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Add correlation ID if not present
	if _, exists := request["correlation_id"]; !exists {
		request["correlation_id"] = fmt.Sprintf("proxy-%d", time.Now().UnixNano())
	}

	// Publish to Redis
	if err := sp.redisPublisher("ai-requests", request); err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish message: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"status":         "queued",
		"correlation_id": request["correlation_id"],
		"message":        "Request queued successfully. Monitor 'ai-responses' channel for results.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (sp *ServiceProxy) Stop() error {
	if sp.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return sp.server.Shutdown(ctx)
	}
	return nil
}