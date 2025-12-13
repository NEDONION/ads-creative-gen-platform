#!/bin/bash

# Script to start the ads-creative-gen-platform server

echo "🚀 Starting Ads Creative Generation Platform..."

# Check if the server is already running
if lsof -Pi :4000 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Server is already running on port 4000!"
    exit 1
fi

# Start the server in the background and redirect output to a log file
nohup go run main.go > server.log 2>&1 &

# Get the PID of the background process
SERVER_PID=$!

if [ $? -eq 0 ]; then
    echo "✅ Server started successfully!"
    echo "📊 Server PID: $SERVER_PID"
    echo "🌐 Server is running on http://localhost:4000"
    echo "📝 Logs are being written to server.log"
else
    echo "❌ Failed to start server"
    exit 1
fi

# Wait a moment for the server to initialize
sleep 2

# Check if the server is actually listening on port 4000
if lsof -Pi :4000 -sTCP:LISTEN -t >/dev/null ; then
    echo "✅ Server confirmed to be listening on port 4000"
else
    echo "❌ Server might not have started properly. Check server.log for details."
    exit 1
fi