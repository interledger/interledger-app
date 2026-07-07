

testci: 
	@echo "Running tests"
	go test -p 2 ./...
	@echo "Tests complete."

test: testenvup
	@echo "Running tests"
	go test -p 2 ./...
	@echo "Not destroying test env as you may run tests again."
	@echo "Tests complete."

testenvup:
	@echo "Starting test environment"
	docker compose -f ./docker-compose-dev.yml up -d
	@echo "Test environment started"

testenvdown:
	@echo "Destroy test environment"
	docker compose -f ./docker-compose-dev.yml down
	@echo "Test environment destoyed"

format:
	@echo "Formatting"
	go fmt ./...
	@echo "Formate complete"

lint:
	@echo "Linting"
	golangci-lint run --go=1.17 ./...
	@echo "Lint complete"

