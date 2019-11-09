package discovery

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.etcd.io/etcd/clientv3"

	"gopkg.qsoa.cloud/service"
)

const (
	TypeService Type = "service"
	TypeMySql   Type = "mysql"
)

var (
	etcdClient    *clientv3.Client
	etcdClientMtx sync.Mutex
)

type Type string

func Watch(ctx context.Context, targetType Type, target string, cb func(instances []Instance)) error {
	etcdClient, err := getEtcd()
	if err != nil {
		return err
	}

	etcdPath := GetPath(targetType, target)
	resp, err := etcdClient.Get(ctx, etcdPath, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	maxRev := int64(0)
	instances := map[string]string{}
	for _, kv := range resp.Kvs {
		if kv.ModRevision > maxRev {
			maxRev = kv.ModRevision
		}
		instances[strings.TrimPrefix(string(kv.Key), etcdPath)] = string(kv.Value)
	}

	cb(mapToInstances(instances))

	ch := etcdClient.Watch(ctx, GetPath(targetType, target), clientv3.WithRev(maxRev), clientv3.WithPrefix())
	for {
		evs, ok := <-ch
		if !ok {
			break
		}

		for _, ev := range evs.Events {
			if ev.Type == clientv3.EventTypeDelete {
				delete(instances, strings.TrimPrefix(string(ev.Kv.Key), etcdPath))
				continue
			}
			instances[strings.TrimPrefix(string(ev.Kv.Key), etcdPath)] = string(ev.Kv.Value)
		}

		cb(mapToInstances(instances))
	}

	return nil
}

func GetPath(targetType Type, id string) string {
	return fmt.Sprintf("/discovery/%s/%s/%s/%s/", service.GetProject(), service.GetEnv(), targetType, id)
}

func getEtcd() (*clientv3.Client, error) {
	etcdClientMtx.Lock()
	defer etcdClientMtx.Unlock()

	if etcdClient != nil {
		return etcdClient, nil
	}

	addr, user, password := service.GetDiscovery()

	return clientv3.New(clientv3.Config{
		Endpoints: strings.Split(addr, ";"),
		Username:  user,
		Password:  password,
	})
}

func mapToInstances(m map[string]string) []Instance {
	res := make([]Instance, 0, len(m))

	for host, v := range m {
		res = append(res, Instance{
			Addr:   host,
			Status: InstanceStatusReady,
			Value:  v,
		})
	}

	return res
}
