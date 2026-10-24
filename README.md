# qSOA cloud service libraries

## Install

### Base library
`go get -u gopkg.qsoa.cloud/service`

### Library for gRPC services
`go get -u gopkg.qsoa.cloud/service/qgrpc`

### Library for HTTP services
`go get -u gopkg.qsoa.cloud/service/qhttp`

### Cloud MySql client
`go get -u gopkg.qsoa.cloud/service/qmysql`

### Cloud ClickHouse client
`go get -u gopkg.qsoa.cloud/service/qclickhouse`

## Usage
```go
package main

import (
	"database/sql"
	"log"

	"gopkg.qsoa.cloud/service"
	"gopkg.qsoa.cloud/service/example/grpc"
	"gopkg.qsoa.cloud/service/example/grpc/pb"
	"gopkg.qsoa.cloud/service/example/http"
	"gopkg.qsoa.cloud/service/qgrpc"
	"gopkg.qsoa.cloud/service/qhttp"
	_ "gopkg.qsoa.cloud/service/qmysql"
)

func main() {
	// Prepare gRpc client
	conn, err := qgrpc.Dial("qcloud://" + service.GetService() + "/")
	if err != nil {
		log.Fatalf("Cannot dial grpc: %v", err)
	}
	defer conn.Close()

	grpcClient := pb.NewTestClient(conn)

	// Prepare mysql connection
	db, err := sql.Open("qmysql", "example_db")
	if err != nil {
		log.Fatalf("Cannot open mysql database: %v", err)
	}
	defer db.Close()

	// Provide HTTP service
	qhttp.Handle("/", http.New(grpcClient, db))

	// Provide gRPC service
	pb.RegisterTestServer(qgrpc.GetServer(), grpc.Server{})

	// Run service
	service.Run()
}

```

## MySQL proxy
With `-q_mysql_sock` (runner passes `unix:///qcloud/mysql_proxy.sock`), `qmysql` connects through runner's MySQL proxy without a password; without it, `qmysql` keeps using `-q_mysql_addr` and `GetDsn`. `qmysql` traces every SQL call.

## ClickHouse
`qclickhouse.Open(ctx, "events")` returns a clickhouse-go v2 `driver.Conn` for an attached ClickHouse database. It connects through runner's socket from `-q_clickhouse_sock` (runner passes `unix:///qcloud/clickhouse.sock`) without a password and pings the database; without the flag it returns `qclickhouse.ErrNotAttached`. The Conn traces every query, select, exec and batch. A context deadline cancels the query but is not sent to ClickHouse as `max_execution_time`.

## Outbound HTTP
Importing the library makes `http.DefaultClient` trace outbound requests. For another client, `qhttp.Transport()` wraps the current `http.DefaultTransport` with a span per request:
```go
client := &http.Client{Transport: qhttp.Transport()}
```

## gRPC gateway
The `qgrpc` server and clients from `qgrpc.Dial` accept and send messages up to 16 MiB. A call that came through a gRPC gateway carries `x-qcloud-via: gateway` and, when the client presented a certificate that the gateway verified, that certificate in `x-client-cert`. `qgrpc.GatewayClientCert(ctx)` returns the certificate only for such calls; for calls from other services or without a certificate it returns false.
