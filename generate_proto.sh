#!/bin/bash

set -e  # Остановить при ошибке

echo "Checking dependencies..."

# Проверка protoc
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc not found. Please install protobuf package."
    exit 1
fi

# Проверка Go плагинов
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Error: protoc-gen-go not found. Please run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Error: protoc-gen-go-grpc not found. Please run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# Проверка Python инструментов
if ! python -c "import grpc_tools.protoc" &> /dev/null; then
    echo "Error: grpc_tools not found. Please run: pip install grpcio-tools"
    exit 1
fi

echo "All dependencies found. Generating proto files..."

echo "Generating Go protobuf files..."
cd go-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/videoproc.proto

echo "Generating Python protobuf files..."
cd ../py-processor
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. proto/videoproc.proto

echo "Proto files generated successfully!"