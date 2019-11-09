package grpc

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/resolver"

	"gopkg.qsoa.cloud/service"
)

var grpcServer *grpc.Server

func init() {
	resolver.Register(&qBuilder{})
}

func Init(opts ...grpc.ServerOption) {
	grpcServer = grpc.NewServer(append([]grpc.ServerOption{
		grpc.UnaryInterceptor(ServerUnaryInterceptor),
		grpc.StreamInterceptor(ServerStreamInterceptor),
	}, opts...)...)
}

func GetServer() *grpc.Server {
	if grpcServer == nil {
		Init()
	}
	return grpcServer
}

func Run() {
	service.Run()

	lis, err := net.Listen("tcp", service.GetListenAddr())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	s := GetServer()
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target, append([]grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true),
		),
		grpc.WithUnaryInterceptor(ClientUnaryInterceptor),
	}, opts...)...)
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
	return handler(srv, ss)
}

func ClientUnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	span := opentracing.SpanFromContext(ctx)

	md := metadata.New(nil)
	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md)); err != nil {
		return err
	}

	mdKv := make([]string, 0, md.Len()*2)
	for k, v := range md {
		mdKv = append(mdKv, k, v[0])
	}
	return invoker(metadata.AppendToOutgoingContext(ctx, mdKv...), method, req, reply, cc, opts...)
}
