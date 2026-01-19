.PHONY: lint format

lint:
	golangci-lint run ./...

format:
	gofmt -w . && goimports -w .