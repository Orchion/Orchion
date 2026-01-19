proto:
	protoc -I ../shared/proto \
		--go_out=internal/proto --go_opt=paths=source_relative \
		--go-grpc_out=internal/proto --go-grpc_opt=paths=source_relative \
		../shared/proto/v1/orchestrator.proto