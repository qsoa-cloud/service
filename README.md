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