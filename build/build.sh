#!/usr/bin/env bash

set -o nounset
set -o errexit

echo "Building:"
go build -o bin/ ./...
echo "Done"