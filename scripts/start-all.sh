#!/bin/bash

# Start Both Python Flask + Go WebSocket Servers
# Run this from the root algo directory

echo "🚀 Starting ALGO Application Stack..."
echo "======================================="
echo ""

# Check for port conflicts and auto-cleanup if needed
echo "🔍 Checking for port conflicts..."
if lsof -ti:8080 >/dev/null 2>&1 || lsof -ti:5000 >/dev/null 2>&1; then
    echo "🧹 Ports in use - running automatic cleanup..."
    ./scripts/cleanup.sh
    echo "✅ Cleanup complete - continuing startup..."
    echo ""
fi

echo "✅ Ports 8080 and 5000 are available"
echo ""

# Function to kill all processes on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down servers..."
    if [ ! -z "$GO_PID" ]; then
        kill $GO_PID 2>/dev/null
        echo "✅ Go WebSocket server stopped"
    fi
    if [ ! -z "$PYTHON_PID" ]; then
        kill $PYTHON_PID 2>/dev/null
        echo "✅ Python Flask server stopped"
    fi
    echo "👋 All servers stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

echo "1️⃣ Starting Migrated Go Backend Server..."
echo "   📡 Port: 8080"
echo "   🚀 Handling Web, API, and WebSockets"
echo "   🔐 Google OAuth authentication"
echo ""

# Start Go server
cd app/src
./start-go-server.sh &
GO_PID=$!
cd ../..

echo ""
echo "✅ Go server is running!"
echo "======================================="
echo "🌐 App URL:       http://localhost:8080"
echo "🔌 WebSocket:     ws://localhost:8080/ws"
echo "📊 Health Check:  http://localhost:8080/health"
echo ""
echo "💡 Press Ctrl+C to stop the server"
echo ""

# Wait for process
wait
