# Docker App Builder - Complete Guide

The Crush Docker App Builder enables creating, building, and running applications in Docker containers directly through chat commands.

## Quick Start

### 1. Create a New App
```bash
# React application
docker_app_builder create_project my-react-app react

# Node.js API  
docker_app_builder create_project my-api nodejs

# Python FastAPI
docker_app_builder create_project my-python-api python

# Go web server
docker_app_builder create_project my-go-app go
```

### 2. Build the App
```bash
docker_app_builder build my-react-app
```

### 3. Run the App
```bash
# Run on default port 3000
docker_app_builder run my-react-app

# Run on custom port with environment variables
docker_app_builder run my-react-app port:8080 environment:{"NODE_ENV":"production"}
```

### 4. Manage Apps
```bash
# List all running apps
docker_app_builder list

# Stop an app
docker_app_builder stop my-react-app
```

## Project Templates

### React Application
- **Features**: Modern React with hooks, live reload, responsive design
- **Build**: Multi-stage build with nginx serving
- **Port**: 3000
- **Files**: package.json, src/App.js, public/index.html, Dockerfile, nginx.conf

### Node.js/Express API
- **Features**: Express server with REST endpoints, health checks
- **Build**: Single-stage Node.js alpine 
- **Port**: 3000
- **Files**: package.json, index.js, Dockerfile

### Python/FastAPI API  
- **Features**: FastAPI with auto-documentation, async support
- **Build**: Python slim with pip requirements
- **Port**: 3000
- **Files**: requirements.txt, main.py, Dockerfile

### Go/Gin Web Server
- **Features**: High-performance Gin server with JSON APIs
- **Build**: Multi-stage build with alpine runtime
- **Port**: 3000
- **Files**: go.mod, main.go, Dockerfile

## Advanced Usage

### Custom Files
```json
{
  "action": "create_project",
  "project_name": "custom-app", 
  "project_type": "nodejs",
  "files": {
    "routes/users.js": "// Custom route file content",
    "config/database.js": "// Database configuration"
  }
}
```

### Environment Variables
```json
{
  "action": "run",
  "project_name": "my-app",
  "environment": {
    "NODE_ENV": "production",
    "DATABASE_URL": "postgresql://localhost:5432/mydb"
  }
}
```

### Custom Commands
```json
{
  "action": "run",
  "project_name": "my-app", 
  "command": "npm run dev"
}
```

## Web Interface Integration

The web chat interface includes Docker-specific features:

- **Docker Status Indicator**: Shows if Docker is available
- **Enhanced Prompts**: Includes Docker app builder examples
- **Real-time API**: Direct integration with Docker tool
- **Error Handling**: Visual feedback for Docker operations

### Example Chat Commands

- "Create a React app using Docker: docker_app_builder create_project my-react-app react"
- "Build a Node.js API: docker_app_builder create_project my-api nodejs"
- "Build and run my Docker app: docker_app_builder build my-app"

## Security & Permissions

All Docker operations go through Crush's permission system:

- Permission requests for Docker operations
- Project isolation in `/tmp/crush-apps/`
- Container naming with `crush-app-` prefix
- Secure command execution

## Troubleshooting

### Docker Not Available
- Ensure Docker is installed and running
- Check Docker daemon status: `docker --version`
- Verify Docker socket access: `/var/run/docker.sock`

### Build Failures
- Check project files are correctly generated
- Verify Dockerfile syntax
- Ensure base images are accessible

### Runtime Issues
- Check port availability (default: 3000)
- Verify container logs: `docker logs crush-app-PROJECT-instance`
- Check Docker resource limits

## API Reference

### Direct Tool Usage

All operations can be called directly via the Docker tool:

```go
dockerTool := tools.NewDockerTool(permissions)
result, err := dockerTool.Run(ctx, tools.ToolCall{
    Name: "docker_app_builder",
    Input: `{"action": "create_project", "project_name": "test", "project_type": "react"}`,
})
```

### Web API Endpoints

- `POST /api/docker` - Execute Docker operations
- `GET /api/health` - Check Docker availability
- `POST /api/chat` - Send Docker commands via chat

This completes the Docker-in-Docker app builder integration with Crush!