package qgrpc

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
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
		grpc.WithStreamInterceptor(ClientStreamInterceptor),
	}

	return grpc.Dial(target, append(opts, addOpts...)...)
}

func ClientUnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("Invoke %s from %s", method, cc.Target()))
	defer span.Finish()

	err := invoker(ctxWithQData(ctx, cc), method, req, reply, cc, opts...)
	if err != nil {
		span.SetTag("error", nil)
		span.LogFields(log.Error(err))
	}

	return err
}

func ClientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("Invoke %s from %s", method, cc.Target()))
	defer span.Finish()

	stream, err := streamer(ctxWithQData(ctx, cc), desc, cc, method, opts...)
	if err != nil {
		span.SetTag("error", nil)
		span.LogFields(log.Error(err))
	}

	return stream, err
}

func ctxWithQData(ctx context.Context, cc *grpc.ClientConn) context.Context {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("grpc", nil)

	md := metadata.New(nil)
	_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md))

	md.Set("x-qcloud-target", cc.Target())

	var pairs []string
	for k, values := range md {
		for _, v := range values {
			pairs = append(pairs, k, v)
		}
	}

	return metadata.AppendToOutgoingContext(ctx, pairs...)
}
