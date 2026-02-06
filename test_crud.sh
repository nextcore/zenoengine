#!/bin/bash

# Setup env
export APP_ENV=development
export DB_DRIVER=sqlite
export DB_NAME=:memory:

# Start server in background
echo "Starting server..."
go run cmd/zeno/zeno.go run src/tutorial/crudapi/main.zl &
PID=$!

# Wait for server to start
sleep 15

echo "Testing GET /api/todos (Empty)..."
OUT=$(curl -s http://localhost:8080/api/todos)
echo "Output: $OUT"

echo "Testing POST /api/todos..."
OUT=$(curl -s -X POST -H "Content-Type: application/json" -d '{"title":"Buy Milk"}' http://localhost:8080/api/todos)
echo "Output: $OUT"

# Kill server
kill $PID
