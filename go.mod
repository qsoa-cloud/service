module gopkg.qsoa.cloud/service

go 1.25.8

require (
	github.com/go-sql-driver/mysql v1.10.1
	github.com/gogo/protobuf v1.3.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.24.1
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	gopkg.qsoa.cloud/grpccodec v0.1.0
	gopkg.qsoa.cloud/tracer v1.0.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.71.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688 // indirect
)

replace gopkg.qsoa.cloud/grpccodec => ../grpccodec
