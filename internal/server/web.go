package server

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed web/build/*
var webFS embed.FS

type WebServer struct {
	port int
}

func NewWebServer(port int) *WebServer {
	return &WebServer{port: port}
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

	// API routes for future Crush backend integration
	http.HandleFunc("/api/", s.handleAPI)

	slog.Info("Starting web server", "port", s.port, "url", fmt.Sprintf("http://localhost:%d", s.port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *WebServer) handleAPI(w http.ResponseWriter, r *http.Request) {
	// Security headers for API
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	
	// CORS headers for development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For now, return a placeholder response
	// In the future, this would integrate with the existing Crush backend
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Crush API - Coming Soon", "status": "placeholder"}`))
}