package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/llm/agent"
	"github.com/charmbracelet/crush/internal/llm/tools"
	"github.com/charmbracelet/crush/internal/permission"
	"github.com/charmbracelet/crush/internal/session"
)

//go:embed web/build/*
var webFS embed.FS

type WebServer struct {
	port        int
	agent       agent.Service
	sessions    session.Service
	permissions permission.Service
}

func NewWebServer(port int, agentService agent.Service, sessions session.Service, permissions permission.Service) *WebServer {
	return &WebServer{
		port:        port,
		agent:       agentService,
		sessions:    sessions,
		permissions: permissions,
	}
}

func (s *WebServer) Start() error {
	// Serve static files from embedded web build
	webBuildFS, err := fs.Sub(webFS, "web/build")
	if err != nil {
		return fmt.Errorf("failed to create web filesystem: %w", err)
	}

	// Create file server for static assets
	fileServer := http.FileServer(http.FS(webBuildFS))

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		path := r.URL.Path
		
		// If it's a static asset, serve it directly
		if strings.HasPrefix(path, "/static/") || 
		   strings.HasSuffix(path, ".js") || 
		   strings.HasSuffix(path, ".css") || 
		   strings.HasSuffix(path, ".ico") || 
		   strings.HasSuffix(path, ".png") || 
		   strings.HasSuffix(path, ".jpg") || 
		   strings.HasSuffix(path, ".svg") {
			fileServer.ServeHTTP(w, r)
			return
		}

		// For all other routes, serve index.html (SPA routing)
		indexFile, err := webBuildFS.Open("index.html")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer indexFile.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		content, err := io.ReadAll(indexFile)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		w.Write(content)
	})

	// API routes
	http.HandleFunc("/api/chat", s.handleChat)
	http.HandleFunc("/api/docker", s.handleDocker)
	http.HandleFunc("/api/sessions", s.handleSessions)
	http.HandleFunc("/api/health", s.handleHealth)

	slog.Info("Starting web server", "port", s.port, "url", fmt.Sprintf("http://localhost:%d", s.port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

// Chat API endpoint
func (s *WebServer) handleChat(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var chatReq ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create or get session
	sessionID := chatReq.SessionID
	if sessionID == "" {
		sessionID = "web-session-" + fmt.Sprintf("%d", time.Now().Unix())
	}

	// Create context
	ctx := context.Background()
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, sessionID)

	// Send message to agent
	eventChan, err := s.agent.Run(ctx, sessionID, chatReq.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Agent error: %v", err), http.StatusInternalServerError)
		return
	}

	// Collect response from event stream
	var responseContent string
	for event := range eventChan {
		if event.Error != nil {
			http.Error(w, fmt.Sprintf("Agent error: %v", event.Error), http.StatusInternalServerError)
			return
		}
		if event.Type == agent.AgentEventTypeResponse {
			responseContent = event.Message.Content().String()
			break
		}
	}

	chatResp := ChatResponse{
		SessionID: sessionID,
		Response:  responseContent,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

// Docker API endpoint
func (s *WebServer) handleDocker(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var dockerReq DockerRequest
	if err := json.NewDecoder(r.Body).Decode(&dockerReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create Docker tool call
	sessionID := dockerReq.SessionID
	if sessionID == "" {
		sessionID = "web-session-" + fmt.Sprintf("%d", time.Now().Unix())
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, sessionID)

	// Create Docker tool
	dockerTool := tools.NewDockerTool(s.permissions)
	
	// Execute Docker command
	toolCall := tools.ToolCall{
		ID:    fmt.Sprintf("docker-%d", time.Now().UnixNano()),
		Name:  tools.DockerToolName,
		Input: string(dockerReq.Params),
	}

	toolResponse, err := dockerTool.Run(ctx, toolCall)
	if err != nil {
		http.Error(w, fmt.Sprintf("Docker tool error: %v", err), http.StatusInternalServerError)
		return
	}

	dockerResp := DockerResponse{
		SessionID: sessionID,
		Success:   !toolResponse.IsError,
		Message:   toolResponse.Content,
		Metadata:  toolResponse.Metadata,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dockerResp)
}

// Sessions API endpoint
func (s *WebServer) handleSessions(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	switch r.Method {
	case "GET":
		// List sessions
		sessions, err := s.sessions.List(context.Background())
		if err != nil {
			http.Error(w, fmt.Sprintf("Error listing sessions: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions": sessions,
		})

	case "POST":
		// Create new session
		var req CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		sessionData := session.Session{
			ID:    fmt.Sprintf("web-session-%d", time.Now().Unix()),
			Title: req.Name,
		}

		_, err := s.sessions.Save(context.Background(), sessionData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating session: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessionData)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Health check endpoint
func (s *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	health := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services: map[string]string{
			"agent":       "active",
			"sessions":    "active",
			"permissions": "active",
		},
	}

	// Check if Docker is available
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		health.Services["docker"] = "available"
	} else {
		health.Services["docker"] = "unavailable"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *WebServer) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// Request/Response types
type ChatRequest struct {
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message"`
}

type ChatResponse struct {
	SessionID string    `json:"session_id"`
	Response  string    `json:"response"`
	Timestamp time.Time `json:"timestamp"`
}

type DockerRequest struct {
	SessionID string          `json:"session_id,omitempty"`
	Params    json.RawMessage `json:"params"`
}

type DockerResponse struct {
	SessionID string    `json:"session_id"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Metadata  string    `json:"metadata,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateSessionRequest struct {
	Name string `json:"name"`
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}