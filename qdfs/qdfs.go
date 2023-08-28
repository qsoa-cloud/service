package qdfs

import (
	"flag"
	"fmt"
	"sync"

	"google.golang.org/grpc"

	"gopkg.qsoa.cloud/service/qdfs/internal/dfspb"
)

var (
	sockAddr     = flag.String("q_dfs_sock", "", "DFS socket")
	dfsClient    dfspb.DfsClient
	dfsClientMtx sync.Mutex
)

func GetFs(bucket string) (*Dfs, error) {
	client, err := getDfsClient()
	if err != nil {
		return nil, err
	}

	return &Dfs{bucket, client}, nil
}

func getDfsClient() (dfspb.DfsClient, error) {
	dfsClientMtx.Lock()
	defer dfsClientMtx.Unlock()

	if dfsClient == nil {
		flag.Parse()

		cc, err := grpc.Dial(*sockAddr, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("cannot connect to dfs server: %v", err)
		}

		dfsClient = dfspb.NewDfsClient(cc)
	}

	return dfsClient, nil
}
