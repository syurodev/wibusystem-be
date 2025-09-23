module wibusystem/pkg/grpc

go 1.25.1

require (
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
	wibusystem/pkg/common v0.0.0
)

replace wibusystem/pkg/common => ../common

require (
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
)
