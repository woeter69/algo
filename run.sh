#!/bin/bash

# ALGO - Development Runner
# This script starts the Go backend and the frontend asset watcher (Tailwind + TS)

# Function to kill background processes on exit
cleanup() {
    echo ""
    echo "🛑 Stopping all services..."
    # Kill the background jobs
    kill $(jobs -p) 2>/dev/null
    # Run the existing cleanup script for ports
    ./scripts/cleanup.sh
    echo "✅ Done."
    exit 0
}

# Trap Ctrl+C (SIGINT) and termination signals
trap cleanup SIGINT SIGTERM

echo "🚀 Starting AlumniGo Unified Stack (Go Backend + Frontend Watcher)..."
echo "========================================================================"

# 1. Start Go Backend
# Note: start-go-server.sh builds and then runs the server
echo "📦 Starting Go Backend (Port 8080)..."
cd app/src
./start-go-server.sh &
cd ../..

# 2. Start Frontend Watcher
# Note: npm run watch runs Tailwind and TypeScript watchers
echo "🎨 Starting Frontend Watcher (Tailwind + TypeScript)..."
npm run watch &

echo ""
echo "✅ Both services are starting up!"
echo "🌐 App URL:       http://localhost:8080"
echo "🔌 WebSocket:     ws://localhost:8080/ws"
echo "========================================================================"
echo "💡 Press Ctrl+C to stop everything"
echo ""

# Keep the script running and wait for background processes
wait
