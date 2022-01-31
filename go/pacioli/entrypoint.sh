#!/bin/bash

function usage {
    echo "usage: entrypoint.sh [test] [lint] [init] [start]"
    echo "  test	run go test ./..."
    echo "  lint	run go vet ./..."
    echo "  migrate	run go run main.go migrate"
    echo "  start	run go run main.go start"
    exit 1
}

if [[ $1 = "test" ]]; then
	echo "Running go test ./..."
	go test ./...
elif [[ $1 = "lint" ]]; then
	echo "Running go vet ./..."
	go vet ./...
elif [[ $1 = "migrate" ]]; then
	echo "Running migrations ./..."
	go run main.go migrate
elif [[ $1 = "start" ]]; then
	echo "Running go vet ./..."
	go run main.go start
elif [[ -z $1 ]]; then
	usage
else
	echo "Unknown command: $1"
fi