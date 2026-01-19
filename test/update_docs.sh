#!/bin/bash

export PATH=$PATH:$(go env GOPATH)/bin

if ! command -v swag &> /dev/null; then
    echo "swag could not be found, installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

echo "Tidying dependencies..."
go mod tidy

echo "Generating Swagger documentation..."
swag init -g cmd/api/main.go --parseDependency --parseInternal
