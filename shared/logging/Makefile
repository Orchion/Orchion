.PHONY: lint format test test-coverage test-coverage-threshold

# Coverage threshold (95% for production code)
COVERAGE_THRESHOLD := 95

lint:
	golangci-lint run ./...

format:
	gofmt -w . && goimports -w .

test:
	go test ./...

test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-coverage-threshold:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total | awk '{print "Coverage: " $$3}'
	@go tool cover -func=coverage.out | grep total | awk '{gsub(/%/, "", $$3); if ($$3 < $(COVERAGE_THRESHOLD)) {print "❌ Coverage below $(COVERAGE_THRESHOLD)% threshold: " $$3 "%"; exit 1} else {print "✅ Coverage meets $(COVERAGE_THRESHOLD)% threshold: " $$3 "%"}}'