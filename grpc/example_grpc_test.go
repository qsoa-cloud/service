//go:generate protoc -I pb --go_out=plugins=grpc:pb pb/service.proto
package grpc_test

import (
	"context"

	"gopkg.qsoa.cloud/service/grpc"
	"gopkg.qsoa.cloud/service/grpc/pb"
)

type Server struct {
}

func (s *Server) Sum(ctx context.Context, r *pb.SumReq) (*pb.SumResp, error) {
	return &pb.SumResp{
		Sum: r.N1 + r.N2,
	}, nil
}

func ExampleRun() {
	pb.RegisterTestServer(grpc.GetServer(), &Server{})

	grpc.Run()

	// Output: test
}
