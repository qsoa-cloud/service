package grpc

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&qBuilder{})
}

func Dial(target string, addOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true),
		),
		grpc.WithUnaryInterceptor(ClientUnaryInterceptor),
		//grpc.WithStreamInterceptor(ClientStreamInterceptor),
	}

	return grpc.Dial(target, append(opts, addOpts...)...)
}

func ClientUnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	span := opentracing.SpanFromContext(ctx)

	md := metadata.New(nil)
	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md)); err != nil {
		return err
	}

	md.Set("x-qcloud-target", cc.Target())

	mdKv := make([]string, 0, md.Len()*2)
	for k, v := range md {
		mdKv = append(mdKv, k, v[0])
	}

	return invoker(metadata.AppendToOutgoingContext(ctx, mdKv...), method, req, reply, cc, opts...)
}

//func ClientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

//}
