#!/usr/bin/env sh

set -o nounset

echo "Running tests:"
echo "" > coverage.out
for d in $(go list ./...); do
    go test -timeout 10s -race -coverprofile=profile.out -covermode=atomic "$d"
    if [ -f profile.out ]; then
        cat profile.out >> coverage.out
        rm profile.out
    fi
done

echo "Checking gofmt:"
for d in $(go list ./...); do
    go fmt "$d"
done

echo "Checking lint:"
golangci-lint run
