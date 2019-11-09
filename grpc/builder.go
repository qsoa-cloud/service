package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc/resolver"

	"gopkg.qsoa.cloud/service/discovery"
)

type qBuilder struct {
}

func (b *qBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	r := &qResolver{
		target: target,
		cc:     cc,
	}

	go r.start()

	return r, nil
}

func (b *qBuilder) Scheme() string {
	return "qcloud"
}

type qResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *qResolver) ResolveNow(resolver.ResolveNowOption) {}

func (r *qResolver) Close() {}

func (r *qResolver) start() {
	if err := discovery.Watch(context.Background(), discovery.TypeService, r.target.Authority, func(instances []discovery.Instance) {
		addrs := make([]resolver.Address, 0, len(instances))
		for _, instance := range instances {
			if instance.Status == discovery.InstanceStatusReady {
				addrs = append(addrs, resolver.Address{
					Type: resolver.Backend,
					Addr: instance.Addr,
				})
			}
		}

		r.cc.UpdateState(resolver.State{
			Addresses: addrs,
		})
	}); err != nil {
		log.Printf("Cannot watch discovery: %s", err.Error())
	}
}
