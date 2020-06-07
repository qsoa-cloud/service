package main

import (
	"log"

	"gopkg.qsoa.cloud/service"
	"gopkg.qsoa.cloud/service/example/grpc"
	"gopkg.qsoa.cloud/service/example/grpc/pb"
	"gopkg.qsoa.cloud/service/example/http"
	qgrpc "gopkg.qsoa.cloud/service/grpc"
)

func main() {
	// Provides HTTP service
	service.HandleHttp("/", &http.Handler{})

	// Provide gRPC service
	service.InitGrpcServer()
	pb.RegisterTestServer(service.GetGrpcServer(), &grpc.Server{})

	// Prepare gRPC client
	var client pb.TestClient
	service.OnInit(func() error {
		conn, err := qgrpc.Dial("qcloud://example/")
		if err != nil {
			log.Fatalf("Cannot dial grpc: %v", err)
		}

		client = pb.NewTestClient(conn)

		return nil
	})

	// Run service
	service.Run()
}
