test: 
	@echo "Running tests"
	go test ./...
	@echo "Tests complete."

testci: test

lint:
	@echo "Linting"
	golangci-lint run ./...
	@echo "Lint complete"

