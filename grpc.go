package service

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
)

var (
	grpcServer *grpc.Server
)

func InitGrpcServer(opts ...grpc.ServerOption) {
	if grpcServer != nil {
		panic("gRpc server is already initialized")
	}

	stdOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(ServerUnaryInterceptor),
		grpc.StreamInterceptor(ServerStreamInterceptor),
	}

	grpcServer = grpc.NewServer(append(stdOpts, opts...)...)
}

func GetGrpcServer() *grpc.Server {
	if grpcServer == nil {
		InitGrpcServer()
	}

	return grpcServer
}

func ServerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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

func ServerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
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

func serveGRpc(l net.Listener, wg *sync.WaitGroup) {
	wg.Add(1)

	// Graceful shutdown on Interrupt signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	go func() {
		<-sigC
		grpcServer.GracefulStop()

		wg.Done()
	}()

	go func() {
		if err := grpcServer.Serve(l); err != nil {
			log.Fatalf("Cannot serve gRPC server: %v", err)
		}
	}()
}
