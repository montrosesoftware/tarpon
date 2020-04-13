#!/usr/bin/env bash

set -o nounset
set -o errexit

echo "Building:"
go build -o bin/fpm-server cmd/fpm-server/main.go
echo "Done"