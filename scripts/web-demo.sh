#!/bin/bash

# Crush Web UI Setup and Demo Script
# This script demonstrates how to set up and run the Crush web interface

set -e

echo "🔧 Crush Web Interface Setup"
echo "=============================="
echo

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "web" ]; then
    echo "❌ Error: Please run this script from the Crush project root"
    exit 1
fi

echo "📦 Installing React dependencies..."
cd web
if [ ! -d "node_modules" ]; then
    echo "Installing npm packages..."
    npm install
else
    echo "Dependencies already installed"
fi

echo
echo "🏗️  Building React application..."
npm run build

echo
echo "✅ Build complete!"
echo

cd ..

echo "🚀 Starting Crush web server..."
echo "The web interface will be available at: http://localhost:8080"
echo
echo "Features available:"
echo "  📱 Mobile-responsive design"
echo "  🚀 Interactive onboarding"
echo "  💬 Chat interface (demo)"
echo "  📚 Documentation"
echo "  🔒 Security information"
echo
echo "Press Ctrl+C to stop the server"
echo

# Start the web server
./crush web --port 8080