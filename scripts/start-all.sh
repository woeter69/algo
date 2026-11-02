#!/bin/bash

# Start AlumniGo Go Server
# Run this from the root algo directory

echo "🚀 Starting AlumniGo Go Server..."
echo "==================================="
echo ""

# Check for port conflicts and auto-cleanup if needed
echo "🔍 Checking for port conflicts..."
if lsof -ti:8080 > /dev/null 2>&1; then
    echo "🧹 Port 8080 in use - running automatic cleanup..."
    ./scripts/cleanup.sh
    echo "✅ Cleanup complete - continuing startup..."
    echo ""
fi

echo "✅ Port 8080 is available"
echo ""

# Function to kill all processes on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down server..."
    if [ ! -z "$GO_PID" ]; then
        kill $GO_PID 2>/dev/null
        echo "✅ Go server stopped"
    fi
    echo "👋 Server stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

echo "1️⃣  Starting Go Backend Server..."
echo "   📡 Port: 8080"
echo "   🚀 Handling Web, API, and WebSockets"
echo "   🔐 Session-based authentication"
echo ""

# Build and start Go server
cd app/src
go build -o websocket-server . 2>&1
if [ $? -ne 0 ]; then
    echo "❌ Go build failed"
    exit 1
fi
./websocket-server &
GO_PID=$!
cd ../..

echo ""
echo "✅ Go server is running! (PID: $GO_PID)"
echo "==================================="
echo "🌐 App URL:       http://localhost:8080"
echo "🔌 WebSocket:     ws://localhost:8080/ws"
echo "📊 Health Check:  http://localhost:8080/health"
echo ""
echo "💡 Press Ctrl+C to stop the server"
echo ""

# Wait for process
wait
