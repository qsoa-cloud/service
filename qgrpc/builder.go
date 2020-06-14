package qgrpc

import (
	"flag"
	"fmt"

	"google.golang.org/grpc/resolver"
)

type qBuilder struct{}

var grpcProxy = flag.String("q_grpc_proxy", "", "gRPC proxy address")

func (b *qBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	flag.Parse()
	if *grpcProxy == "" {
		return nil, fmt.Errorf("no gRpc proxy was provided by -q_grpc_proxy argument")
	}

	r := &qResolver{
		target: target,
		cc:     cc,
	}

	r.cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{
			{Addr: *grpcProxy},
		},
	})

	return r, nil
}

func (b *qBuilder) Scheme() string {
	return "qcloud"
}

type qResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *qResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *qResolver) Close() {}
