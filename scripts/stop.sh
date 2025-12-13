#!/bin/bash

# Script to stop the ads-creative-gen-platform server

echo "🛑 Stopping Ads Creative Generation Platform..."

# Find the process running on port 4000
PID=$(lsof -ti:4000)

if [ -z "$PID" ]; then
    echo "❌ No process found running on port 4000"
    echo "💡 The server might not be running"
    exit 1
else
    echo "🗑️  Found process with PID: $PID"
    
    # Attempt graceful shutdown first
    kill -TERM $PID
    
    # Wait a few seconds to allow graceful shutdown
    sleep 3
    
    # Check if the process is still running
    if kill -0 $PID 2>/dev/null; then
        echo "⏳ Process still running, forcing termination..."
        kill -9 $PID
        sleep 1
    fi
    
    # Verify the process is gone
    if ! kill -0 $PID 2>/dev/null; then
        echo "✅ Server with PID $PID has been stopped"
    else
        echo "❌ Failed to stop server with PID $PID"
        exit 1
    fi
fi

# Also clean up any potential orphaned go processes (though be cautious with this)
echo "📋 Checking for any remaining go run processes..."
GO_RUN_PIDS=$(pgrep -f "go run main.go")

if [ ! -z "$GO_RUN_PIDS" ]; then
    echo "🧹 Cleaning up any remaining go run processes: $GO_RUN_PIDS"
    pkill -f "go run main.go" 2>/dev/null
fi

echo "🎉 Server shutdown complete!"