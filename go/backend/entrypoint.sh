#!/bin/bash

function usage {
    echo "usage: entrypoint.sh [test] [lint]"
    echo "  test	run go test ./..."
    echo "  lint	run go vet ./..."
    exit 1
}

if [[ $1 = "test" ]]; then
	echo "Running go test ./..."
	go test ./...
elif [[ $1 = "lint" ]]; then
	echo "Running go vet ./..."
	go vet ./...
elif [[ -z $1 ]]; then
	usage
else
	echo "Unknown command: $1"
fi