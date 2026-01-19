#!/bin/bash

echo "[1/3] Installing dependencies..."
go mod tidy

export PATH=$PATH:$(go env GOPATH)/bin
if ! command -v swag &> /dev/null; then
    echo "swag could not be found, installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

echo "[2/3] Generating docs..."
./update_docs.sh

echo "[3/3] Starting server..."
echo "Server will be available at http://localhost:8080"
echo "Swagger docs available at http://localhost:8080/swagger/index.html"
go run cmd/api/main.go
