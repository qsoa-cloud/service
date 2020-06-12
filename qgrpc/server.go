package qgrpc

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"gopkg.qsoa.cloud/service"
)

type qGRpc struct {
	server *grpc.Server
}

var (
	gRpcService = &qGRpc{
		server: grpc.NewServer(
			grpc.UnaryInterceptor(serverUnaryInterceptor),
			grpc.StreamInterceptor(serverStreamInterceptor),
		),
	}
)

func init() {
	service.RegisterService(gRpcService)
}

func (s *qGRpc) GetName() string {
	return "grpc"
}

func (s *qGRpc) Serve(l net.Listener, wg *sync.WaitGroup) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)

	go func() {
		<-sigC
		s.server.GracefulStop()

		wg.Done()
	}()

	go func() {
		if err := s.server.Serve(l); err != nil {
			log.Fatalf("Cannot serve gRPC server: %v", err)
		}
	}()
}

func GetServer() *grpc.Server {
	return gRpcService.server
}

func serverUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	sCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md),
	)
	if err != nil {
		return nil, err
	}

	span := opentracing.StartSpan(info.FullMethod, opentracing.ChildOf(sCtx))
	defer span.Finish()

	span.SetTag("grpc", nil)

	return handler(opentracing.ContextWithSpan(ctx, span), req)
}

func serverStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		md = metadata.New(nil)
	}

	sCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md),
	)
	if err != nil {
		return err
	}

	span := opentracing.StartSpan(info.FullMethod, opentracing.ChildOf(sCtx))
	defer span.Finish()

	span.SetTag("grpc", nil)

	return handler(srv, ss)
}
