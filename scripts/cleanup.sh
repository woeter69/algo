#!/bin/bash

# Cleanup script to kill any leftover server processes

echo "🧹 Cleaning up leftover server processes..."

# Kill any Go WebSocket servers on port 8080
if lsof -ti:8080 >/dev/null 2>&1; then
    echo "🔫 Killing Go WebSocket server on port 8080..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null
    echo "✅ Port 8080 freed"
else
    echo "✅ Port 8080 is already free"
fi

# Kill any websocket-server processes
if pgrep -f "websocket-server" >/dev/null 2>&1; then
    echo "🔫 Killing websocket-server processes..."
    pkill -f "websocket-server"
    echo "✅ WebSocket server processes killed"
fi

echo ""
echo "✅ Cleanup complete! Ready to start fresh."
echo "🚀 Now run: ./run.sh"
