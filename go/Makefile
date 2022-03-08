
lint:
	@echo "Linting"
	golangci-lint run ./...
	@echo "Lint complete"

lintci:
	golangci-lint run --issues-exit-code 1 --out-format code-climate | tee gl-code-quality-report.json | jq -r '.[] | "\(.location.path):\(.location.lines.begin) \(.description)"'