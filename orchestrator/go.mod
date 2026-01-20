module github.com/Orchion/Orchion/orchestrator

go 1.21

require (
	github.com/Orchion/Orchion/shared/logging v0.0.0
	google.golang.org/grpc v1.66.3
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240610135401-a8a62080eff3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Orchion/Orchion/shared/logging => ../shared/logging
