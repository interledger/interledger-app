#!/bin/bash

function usage {
    echo "usage: entrypoint.sh [test] [lint]"
    echo "  test	run go test ./..."
    echo "  lint	run go vet ./..."
    echo "  start	run go run main.go start"
    echo "  migrate	run go run main.go migrate"
    exit 1
}

if [[ $1 = "test" ]]; then
	echo "Running go test ./..."
	go test ./...
elif [[ $1 = "lint" ]]; then
	echo "Running go vet ./..."
	go vet ./...
elif [[ $1 = "start" ]]; then
	echo "Running go run main.go start"
	go run main.go start
elif [[ $1 = "migrate" ]]; then
	echo "Running go run main.go migrate"
	go run main.go migrate
elif [[ -z $1 ]]; then
	usage
else
	echo "Unknown command: $1"
fi