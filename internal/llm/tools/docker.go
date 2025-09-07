package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/crush/internal/permission"
)

const DockerToolName = "docker_app_builder"

type DockerAppBuilderParams struct {
	Action      string            `json:"action"`
	ProjectName string            `json:"project_name,omitempty"`
	ProjectType string            `json:"project_type,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
	Command     string            `json:"command,omitempty"`
	Port        string            `json:"port,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

type DockerResponseMetadata struct {
	Action      string `json:"action"`
	ProjectName string `json:"project_name,omitempty"`
	ImageID     string `json:"image_id,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	URL         string `json:"url,omitempty"`
}

type dockerTool struct {
	permissions permission.Service
}

func NewDockerTool(permissions permission.Service) *dockerTool {
	return &dockerTool{permissions: permissions}
}

func (d *dockerTool) Name() string {
	return DockerToolName
}

func (d *dockerTool) Info() ToolInfo {
	return ToolInfo{
		Name:        DockerToolName,
		Description: dockerDescription(),
		Parameters:  dockerProperties(),
		Required:    []string{"action"},
	}
}

func (d *dockerTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var params DockerAppBuilderParams
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Check Docker permission
	sessionID, toolCallID := GetContextValues(ctx)
	permissionRequest := permission.CreatePermissionRequest{
		SessionID:   sessionID,
		ToolCallID:  toolCallID,
		ToolName:    DockerToolName,
		Description: fmt.Sprintf("Docker %s operation", params.Action),
		Action:      params.Action,
		Params:      params,
		Path:        fmt.Sprintf("/tmp/crush-apps/%s", params.ProjectName),
	}
	
	if !d.permissions.Request(permissionRequest) {
		return NewTextErrorResponse("Permission denied for Docker operation"), nil
	}

	// Check if Docker is available
	if err := d.checkDockerAvailable(); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Docker is not available: %v", err)), nil
	}

	switch params.Action {
	case "create_project":
		return d.createProject(ctx, params)
	case "build":
		return d.buildApp(ctx, params)
	case "run":
		return d.runApp(ctx, params)
	case "stop":
		return d.stopApp(ctx, params)
	case "list":
		return d.listContainers(ctx)
	default:
		return NewTextErrorResponse(fmt.Sprintf("Unknown action: %s", params.Action)), nil
	}
}

func (d *dockerTool) checkDockerAvailable() error {
	cmd := exec.Command("docker", "--version")
	return cmd.Run()
}

func (d *dockerTool) createProject(ctx context.Context, params DockerAppBuilderParams) (ToolResponse, error) {
	if params.ProjectName == "" || params.ProjectType == "" {
		return NewTextErrorResponse("project_name and project_type are required for create_project action"), nil
	}

	projectDir := filepath.Join("/tmp", "crush-apps", params.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to create project directory: %v", err)), nil
	}

	// Generate project files based on type
	projectFiles, err := d.generateProjectFiles(params.ProjectType, params.ProjectName)
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to generate project files: %v", err)), nil
	}

	// Add any custom files provided
	for filename, content := range params.Files {
		projectFiles[filename] = content
	}

	// Write all files to the project directory
	for filename, content := range projectFiles {
		filePath := filepath.Join(projectDir, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return NewTextErrorResponse(fmt.Sprintf("Failed to create directory for %s: %v", filename, err)), nil
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return NewTextErrorResponse(fmt.Sprintf("Failed to write file %s: %v", filename, err)), nil
		}
	}

	content := fmt.Sprintf("✅ Project '%s' created successfully!\n\nLocation: %s\nType: %s\nGenerated files: %s\n\nNext steps:\n1. Build the project: {\"action\": \"build\", \"project_name\": \"%s\"}\n2. Run the project: {\"action\": \"run\", \"project_name\": \"%s\"}", 
		params.ProjectName, projectDir, params.ProjectType, strings.Join(getKeys(projectFiles), ", "), params.ProjectName, params.ProjectName)

	metadata := DockerResponseMetadata{
		Action:      "create_project",
		ProjectName: params.ProjectName,
	}

	return WithResponseMetadata(NewTextResponse(content), metadata), nil
}

func (d *dockerTool) buildApp(ctx context.Context, params DockerAppBuilderParams) (ToolResponse, error) {
	if params.ProjectName == "" {
		return NewTextErrorResponse("project_name is required for build action"), nil
	}

	projectDir := filepath.Join("/tmp", "crush-apps", params.ProjectName)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return NewTextErrorResponse(fmt.Sprintf("Project directory %s does not exist. Create the project first using create_project action.", projectDir)), nil
	}

	// Build the Docker image
	imageName := fmt.Sprintf("crush-app-%s", strings.ToLower(params.ProjectName))
	
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageName, projectDir)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("❌ Docker build failed: %v\n\nOutput:\n%s", err, string(output))), nil
	}

	content := fmt.Sprintf("✅ Successfully built Docker image: %s\n\nBuild output:\n%s\n\nNext step: Run the app with {\"action\": \"run\", \"project_name\": \"%s\"}", 
		imageName, string(output), params.ProjectName)

	metadata := DockerResponseMetadata{
		Action:      "build",
		ProjectName: params.ProjectName,
		ImageID:     imageName,
	}

	return WithResponseMetadata(NewTextResponse(content), metadata), nil
}

func (d *dockerTool) runApp(ctx context.Context, params DockerAppBuilderParams) (ToolResponse, error) {
	if params.ProjectName == "" {
		return NewTextErrorResponse("project_name is required for run action"), nil
	}

	imageName := fmt.Sprintf("crush-app-%s", strings.ToLower(params.ProjectName))
	port := params.Port
	if port == "" {
		port = "3000" // Default port
	}

	// Build run command
	containerName := fmt.Sprintf("crush-app-%s-instance", strings.ToLower(params.ProjectName))
	
	// Check if container already exists and remove it
	exec.Command("docker", "rm", "-f", containerName).Run()
	
	runArgs := []string{"run", "-d", "-p", fmt.Sprintf("%s:%s", port, port)}
	
	// Add environment variables
	for key, value := range params.Environment {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	
	// Add container name
	runArgs = append(runArgs, "--name", containerName)
	
	// Add image name
	runArgs = append(runArgs, imageName)
	
	// Add custom command if provided
	if params.Command != "" {
		runArgs = append(runArgs, "sh", "-c", params.Command)
	}

	cmd := exec.CommandContext(ctx, "docker", runArgs...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("❌ Docker run failed: %v\n\nOutput:\n%s", err, string(output))), nil
	}

	containerID := strings.TrimSpace(string(output))
	appURL := fmt.Sprintf("http://localhost:%s", port)
	
	content := fmt.Sprintf("✅ Successfully started container: %s\n\nContainer ID: %s\nApp URL: %s\n\nThe app is now running! You can:\n- Visit %s in your browser\n- Stop it with: {\"action\": \"stop\", \"project_name\": \"%s\"}\n- View logs with: docker logs %s", 
		containerName, containerID, appURL, appURL, params.ProjectName, containerName)

	metadata := DockerResponseMetadata{
		Action:      "run",
		ProjectName: params.ProjectName,
		ContainerID: containerID,
		URL:         appURL,
	}

	return WithResponseMetadata(NewTextResponse(content), metadata), nil
}

func (d *dockerTool) stopApp(ctx context.Context, params DockerAppBuilderParams) (ToolResponse, error) {
	if params.ProjectName == "" {
		return NewTextErrorResponse("project_name is required for stop action"), nil
	}

	containerName := fmt.Sprintf("crush-app-%s-instance", strings.ToLower(params.ProjectName))
	
	// Stop the container
	cmd := exec.CommandContext(ctx, "docker", "stop", containerName)
	output, err := cmd.CombinedOutput()
	
	var content string
	if err != nil {
		content = fmt.Sprintf("⚠️ Container %s was not running or already stopped.\n\nOutput: %s", containerName, string(output))
	} else {
		content = fmt.Sprintf("✅ Successfully stopped container: %s\n\nOutput: %s", containerName, string(output))
	}

	// Remove the container
	exec.CommandContext(ctx, "docker", "rm", containerName).Run()
	content += fmt.Sprintf("\n🗑️ Container %s removed.", containerName)

	metadata := DockerResponseMetadata{
		Action:      "stop",
		ProjectName: params.ProjectName,
	}

	return WithResponseMetadata(NewTextResponse(content), metadata), nil
}

func (d *dockerTool) listContainers(ctx context.Context) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "name=crush-app", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("❌ Failed to list containers: %v\n\nOutput: %s", err, string(output))), nil
	}

	content := fmt.Sprintf("📋 Crush App Containers:\n\n%s\n\nTo interact with these containers:\n- Stop: {\"action\": \"stop\", \"project_name\": \"PROJECT_NAME\"}\n- View logs: docker logs CONTAINER_NAME", string(output))

	metadata := DockerResponseMetadata{
		Action: "list",
	}

	return WithResponseMetadata(NewTextResponse(content), metadata), nil
}

func (d *dockerTool) generateProjectFiles(projectType, projectName string) (map[string]string, error) {
	files := make(map[string]string)
	
	switch projectType {
	case "nodejs", "express":
		files["package.json"] = fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "A Crush-generated Node.js application",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "node index.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}`, projectName)
		
		files["index.js"] = `const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({ 
    message: 'Hello from Crush-generated Node.js app!',
    timestamp: new Date().toISOString(),
    project: process.env.npm_package_name || 'crush-app'
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'healthy', uptime: process.uptime() });
});

app.listen(port, '0.0.0.0', () => {
  console.log('🚀 Server running on port', port);
});`

		files["Dockerfile"] = `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]`

	case "python", "fastapi":
		files["requirements.txt"] = `fastapi==0.104.1
uvicorn==0.24.0`
		
		files["main.py"] = `from fastapi import FastAPI
from datetime import datetime
import uvicorn
import os

app = FastAPI(title="Crush-generated Python API")

@app.get("/")
def read_root():
    return {
        "message": "Hello from Crush-generated Python app!",
        "timestamp": datetime.now().isoformat(),
        "project": os.environ.get("PROJECT_NAME", "crush-app")
    }

@app.get("/health")
def health_check():
    return {"status": "healthy"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=3000)`

		files["Dockerfile"] = `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 3000
CMD ["python", "main.py"]`

	case "go":
		files["go.mod"] = fmt.Sprintf(`module %s

go 1.21

require github.com/gin-gonic/gin v1.9.1`, projectName)
		
		files["main.go"] = `package main

import (
	"net/http"
	"os"
	"time"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Hello from Crush-generated Go app!",
			"timestamp": time.Now().Format(time.RFC3339),
			"project":   getEnv("PROJECT_NAME", "crush-app"),
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	r.Run("0.0.0.0:3000")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}`

		files["Dockerfile"] = `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 3000
CMD ["./main"]`

	case "react":
		files["package.json"] = fmt.Sprintf(`{
  "name": "%s",
  "version": "0.1.0",
  "private": true,
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-scripts": "5.0.1"
  },
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "serve": "npx serve -s build -l 3000"
  },
  "eslintConfig": {
    "extends": ["react-app"]
  },
  "browserslist": {
    "production": [">0.2%%", "not dead", "not op_mini all"],
    "development": ["last 1 chrome version", "last 1 firefox version", "last 1 safari version"]
  }
}`, projectName)

		files["public/index.html"] = fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>%s - Crush Generated App</title>
    <style>
      body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif; }
    </style>
  </head>
  <body>
    <div id="root"></div>
  </body>
</html>`, projectName)

		files["src/index.js"] = `import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<App />);`

		files["src/App.js"] = fmt.Sprintf(`import React, { useState, useEffect } from 'react';

function App() {
  const [time, setTime] = useState(new Date().toLocaleTimeString());
  const [status, setStatus] = useState('Starting...');

  useEffect(() => {
    const timer = setInterval(() => {
      setTime(new Date().toLocaleTimeString());
    }, 1000);

    // Simulate app loading
    setTimeout(() => setStatus('Running'), 1000);

    return () => clearInterval(timer);
  }, []);

  return (
    <div style={{ 
      textAlign: 'center', 
      padding: '50px',
      background: 'linear-gradient(135deg, #667eea 0%%, #764ba2 100%%)',
      minHeight: '100vh',
      color: 'white'
    }}>
      <h1>🚀 %s</h1>
      <p style={{ fontSize: '1.2em', marginBottom: '20px' }}>
        Hello from your Crush-generated React app!
      </p>
      <div style={{ 
        background: 'rgba(255,255,255,0.1)', 
        padding: '20px', 
        borderRadius: '10px',
        display: 'inline-block',
        marginBottom: '20px'
      }}>
        <p><strong>Status:</strong> {status}</p>
        <p><strong>Current time:</strong> {time}</p>
        <p><strong>Built with:</strong> Docker-in-Docker ⚡</p>
      </div>
      <div style={{ marginTop: '30px' }}>
        <h3>🎉 Your app is ready!</h3>
        <p>This React application was generated and containerized by Crush.</p>
      </div>
    </div>
  );
}

export default App;`, projectName)

		files["Dockerfile"] = `# Multi-stage build for React app
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Serve with nginx
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]`

		files["nginx.conf"] = `server {
    listen 3000;
    server_name localhost;
    
    location / {
        root /usr/share/nginx/html;
        index index.html index.htm;
        try_files $uri $uri/ /index.html;
    }
    
    # Enable gzip compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
}`

	default:
		return nil, fmt.Errorf("unsupported project type: %s. Supported types: nodejs, python, go, react, express, fastapi", projectType)
	}

	return files, nil
}

func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func dockerDescription() string {
	return `Docker App Builder - Create, build, and run applications in Docker containers with Docker-in-Docker support.

This tool provides a complete Docker-based application development workflow through chat commands:

## Actions Available:

### create_project
Creates a new application project with scaffolded files:
- **project_name**: Name for the new project (required)
- **project_type**: Type of project (required): nodejs, python, go, react, express, fastapi
- **files**: Optional custom files to add to the project

### build
Builds a Docker image for the project:
- **project_name**: Name of the project to build (required)

### run  
Runs the Docker container:
- **project_name**: Name of the project to run (required)
- **port**: Port to expose (default: 3000)
- **environment**: Environment variables to set
- **command**: Custom command to run in container

### stop
Stops and removes the running container:
- **project_name**: Name of the project to stop (required)

### list
Lists all Crush app containers and their status

## Project Types Supported:

1. **nodejs/express** - Express.js server with REST API endpoints
2. **python/fastapi** - FastAPI server with REST API endpoints
3. **go** - Gin web server with REST API endpoints  
4. **react** - React frontend application with modern UI

## Complete Workflow Example:

1. Create: {"action": "create_project", "project_name": "my-app", "project_type": "react"}
2. Build: {"action": "build", "project_name": "my-app"}
3. Run: {"action": "run", "project_name": "my-app", "port": "3000"}
4. Visit: http://localhost:3000
5. Stop: {"action": "stop", "project_name": "my-app"}

Each project type includes:
- Appropriate Dockerfile for multi-stage builds
- Package management files (package.json, requirements.txt, go.mod)
- Starter application code with health endpoints
- Production-ready configuration

All projects are created in /tmp/crush-apps/ and containers use 'crush-app-' naming.
Docker must be installed and running for this tool to work.`
}

func dockerProperties() map[string]any {
	return map[string]any{
		"action": map[string]any{
			"type":        "string",
			"description": "Action to perform",
			"enum":        []string{"create_project", "build", "run", "stop", "list"},
		},
		"project_name": map[string]any{
			"type":        "string",
			"description": "Name of the project (required for most actions)",
		},
		"project_type": map[string]any{
			"type":        "string",
			"description": "Type of project to create (required for create_project)",
			"enum":        []string{"nodejs", "python", "go", "react", "express", "fastapi"},
		},
		"files": map[string]any{
			"type":        "object",
			"description": "Custom files to add to the project (filename -> content mapping)",
			"additionalProperties": map[string]any{
				"type": "string",
			},
		},
		"command": map[string]any{
			"type":        "string",
			"description": "Custom command to run in the container",
		},
		"port": map[string]any{
			"type":        "string",
			"description": "Port to expose (default: 3000)",
		},
		"environment": map[string]any{
			"type":        "object",
			"description": "Environment variables to set in the container",
			"additionalProperties": map[string]any{
				"type": "string",
			},
		},
	}
}