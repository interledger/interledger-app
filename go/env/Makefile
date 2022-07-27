testci: test

test:
	@echo "Running tests"
	go test ./...
	@echo "Tests complete."

format:
	@echo "Formatting"
	go fmt ./...
	@echo "Formate complete"

lint:
	@echo "Linting"
	golangci-lint --go=1.17 run ./...
	@echo "Lint complete"
