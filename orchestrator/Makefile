proto:
	protoc -I ../shared/proto \
		--go_out=api/v1 --go_opt=paths=source_relative \
		--go-grpc_out=api/v1 --go-grpc_opt=paths=source_relative \
		../shared/proto/v1/orchestrator.proto