#!/bin/bash

echo "Installing dependencies..."
go mod tidy

echo "Generating docs..."
./update_docs.sh

echo "Starting server in background..."
export MOCK_AUTH=true
go run cmd/api/main.go &
PID=$!
sleep 5

echo "Running tests..."

echo "Testing POST /auth/login"
curl -X POST -H "Authorization: Bearer mock-token" -H "Content-Type: application/json" -d '{"role": "admin"}' http://localhost:8080/auth/login
echo "\n"
echo "Testing POST /api/account"
curl -X POST -H "Authorization: Bearer mock-token" -H "Content-Type: application/json" -d '{"name": "test_name", "created_at": "2023-01-01T00:00:00Z"}' http://localhost:8080/api/account
echo "\n"
echo "Testing POST /api/User"
curl -X POST -H "Authorization: Bearer mock-token" -H "Content-Type: application/json" -d '{"uid": "test_uid", "email": "test_email", "name": "test_name", "picture": "test_picture", "roleId": "test_roleId", "settingsId": "test_settingsId", "created_at": "2023-01-01T00:00:00Z", "updated_at": "2023-01-01T00:00:00Z"}' http://localhost:8080/api/User
echo "\n"

echo "Killing server (PID: $PID)..."
kill $PID
