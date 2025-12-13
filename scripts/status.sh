#!/bin/bash

# Script to check the status of the ads-creative-gen-platform server

echo "🔍 Checking status of Ads Creative Generation Platform..."

# Check if the server is running on port 4000
if lsof -Pi :4000 -sTCP:LISTEN -t >/dev/null ; then
    PID=$(lsof -ti:4000)
    echo "✅ Server is RUNNING"
    echo "📊 PID: $PID"
    echo "🌐 Running on: http://localhost:4000"
    
    # Get additional process info
    PROCESS_INFO=$(ps -p $PID -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo ""
        echo "📋 Process Info:"
        echo "$PROCESS_INFO"
    fi
    
    # Test connectivity
    if curl -s http://localhost:4000/health >/dev/null 2>&1; then
        echo ""
        echo "🔗 Health check: SUCCESS"
        HEALTH_RESPONSE=$(curl -s http://localhost:4000/health)
        echo "💬 Health response: $HEALTH_RESPONSE"
    else
        echo ""
        echo "⚠️  Health check: FAILED - Server may be running but not responding"
    fi
else
    echo "🔴 Server is STOPPED or not running on port 4000"
    echo "💡 Use './start.sh' to start the server"
fi