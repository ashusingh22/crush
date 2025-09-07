# Crush Web UI

This directory contains the React-based web interface for Crush, providing a mobile-friendly way to interact with the AI development assistant.

## Features

- 📱 **Mobile-Responsive Design**: Optimized for phones, tablets, and desktops
- 🚀 **Interactive Onboarding**: Step-by-step guide to Crush features and security
- 💬 **Chat Interface**: Real-time conversation with AI assistant
- 📚 **Documentation**: Comprehensive guide to tools and security features
- 🔒 **Security Focused**: Highlights security enhancements and best practices

## Quick Start

1. **Install Dependencies**
   ```bash
   cd web
   npm install
   ```

2. **Development Server**
   ```bash
   npm start
   # Opens http://localhost:3000
   ```

3. **Build for Production**
   ```bash
   npm run build
   # Creates optimized build in build/ directory
   ```

4. **Serve via Crush**
   ```bash
   # After building, serve via Crush web server
   cd ..
   go run . web --port 8080
   # Opens http://localhost:8080
   ```

## Architecture

- **React 18**: Modern React with hooks and functional components
- **Styled Components**: CSS-in-JS for maintainable styling
- **Framer Motion**: Smooth animations and transitions
- **React Router**: Client-side routing for SPA
- **React Icons**: Consistent iconography

## Security Features Highlighted

The web interface emphasizes the security enhancements made to Crush:

1. **YOLO Mode Warnings**: Clear explanations of security implications
2. **Permission System**: Documentation of granular controls
3. **Path Validation**: Explanation of directory traversal protection
4. **Command Sanitization**: Details on safe command execution

## Mobile Optimization

- Touch-friendly interface elements
- Responsive breakpoints for all screen sizes
- Optimized typography and spacing
- Progressive Web App capabilities

## Integration with Crush Backend

The web interface is designed to integrate with the existing Crush backend:

- **API Endpoints**: `/api/*` routes for backend communication
- **Security Headers**: Proper CORS and security configurations
- **Session Management**: Future integration with Crush permissions
- **Real-time Updates**: WebSocket support for live interactions

## Development

The web interface can be developed independently:

```bash
# Development with hot reload
npm start

# Run tests
npm test

# Type checking (if using TypeScript)
npm run type-check

# Linting
npm run lint
```

## Deployment

For production deployment:

1. Build the React app: `npm run build`
2. The built files are embedded into the Go binary
3. Serve via `crush web` command
4. Access at configured port (default: 8080)

## Security Considerations

- All routes are served with security headers
- CORS is configured for development
- Static assets are served securely
- API routes include content-type validation
- Input sanitization on all forms