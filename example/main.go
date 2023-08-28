package main

import (
	"context"
	"database/sql"
	"io"
	"log"
	"os"

	"gopkg.qsoa.cloud/service"
	"gopkg.qsoa.cloud/service/qdfs"
	"gopkg.qsoa.cloud/service/qgrpc"
	"gopkg.qsoa.cloud/service/qhttp"
	_ "gopkg.qsoa.cloud/service/qmysql"

	"example/grpc"
	"example/grpc/pb"
	"example/http"
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

	// Read file from DFS
	fs, err := qdfs.GetFs("example")
	if err != nil {
		log.Fatalf("Cannot get FS from the example bucket: %v", err)
	}
	f, err := fs.OpenFile(context.Background(), "test", os.O_RDONLY)
	if err != nil {
		log.Fatalf("Cannot open file: %v", err)
	}
	if _, err := io.Copy(os.Stderr, f); err != nil {
		log.Fatalf("Cannot copy file to STDERR: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("Cannot close file: %v", err)
	}

	// Provide HTTP service
	qhttp.Handle("/", http.New(grpcClient, db))

	// Provide gRPC service
	pb.RegisterTestServer(qgrpc.GetServer(), grpc.Server{})

	// Run service
	service.Run()
}
