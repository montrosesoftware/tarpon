#!/usr/bin/env sh

set -o nounset
set -o errexit

echo "Building:"
go build -o bin/tarpon cmd/tarpon/main.go
echo "Done"