.PHONY: proto lint format

proto:
	protoc -I ../shared/proto \
		--go_out=api --go_opt=paths=source_relative \
		--go-grpc_out=api --go-grpc_opt=paths=source_relative \
		../shared/proto/v1/orchestrator.proto

lint:
	golangci-lint run ./...

format:
	gofmt -w . && goimports -w .