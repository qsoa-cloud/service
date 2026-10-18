//go:generate protoc -I internal/secretspb --gogofaster_out=plugins=grpc:internal/secretspb internal/secretspb/qsecrets.proto

// Package qsecrets reads the service's secrets. Read them at start: each call reaches the platform.
package qsecrets

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gopkg.qsoa.cloud/service/internal/qconn"
	"gopkg.qsoa.cloud/service/qsecrets/internal/secretspb"
)

var sockAddr = flag.String("q_secrets_sock", "", "Secrets socket")

var ErrNotFound = errors.New("qsecrets: secret not found")

var errNoSocket = errors.New("qsecrets: the runner provides no secrets socket")

var (
	clientMtx sync.Mutex
	client    secretspb.QSecretsClient
)

// Get returns the secret at path for this service and env.
func Get(ctx context.Context, path string) (string, error) {
	c, err := getClient()
	if err != nil {
		return "", err
	}
	resp, err := c.Get(ctx, &secretspb.GetSecretReq{Path: path})
	if status.Code(err) == codes.NotFound {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("qsecrets: %s: %w", path, err)
	}
	return resp.Value, nil
}

func getClient() (secretspb.QSecretsClient, error) {
	clientMtx.Lock()
	defer clientMtx.Unlock()
	if client != nil {
		return client, nil
	}
	flag.Parse()
	if *sockAddr == "" {
		return nil, errNoSocket
	}
	cc, err := qconn.Dial(*sockAddr)
	if err != nil {
		return nil, err
	}
	client = secretspb.NewQSecretsClient(cc)
	return client, nil
}
