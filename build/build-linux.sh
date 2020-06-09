#!/usr/bin/env bash

set -o nounset
set -o errexit

echo "Building Linux 64bit:"
GOOS=linux GOARCH=amd64 go build -o bin/tarpon cmd/tarpon/main.go
echo "Done"