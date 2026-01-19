module github.com/Orchion/Orchion/node-agent

go 1.21

require (
	github.com/Orchion/Orchion/shared/logging v0.0.0
	github.com/google/uuid v1.6.0
	github.com/shirou/gopsutil/v3 v3.24.5
	google.golang.org/grpc v1.66.3
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240610135401-a8a62080eff3 // indirect
)

replace github.com/Orchion/Orchion/shared/logging => ../shared/logging
