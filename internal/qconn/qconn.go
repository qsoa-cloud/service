// Package qconn dials runner transports.
package qconn

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gopkg.qsoa.cloud/grpccodec"
)

// Dial opens a client connection to a runner transport using the gogo-aware codec.
func Dial(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodecV2(grpccodec.Codec)))
}
