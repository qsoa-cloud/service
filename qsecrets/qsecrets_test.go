package qsecrets

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gopkg.qsoa.cloud/service/qsecrets/internal/secretspb"
)

type fake struct {
	secretspb.UnimplementedQSecretsServer
}

func (fake) Get(_ context.Context, r *secretspb.GetSecretReq) (*secretspb.GetSecretResp, error) {
	switch r.Path {
	case "db/password":
		return &secretspb.GetSecretResp{Value: "v"}, nil
	case "missing":
		return nil, status.Error(codes.NotFound, "secret not found")
	}
	return nil, status.Error(codes.Unavailable, "secrets are unavailable")
}

func serve(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "secrets.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	secretspb.RegisterQSecretsServer(s, fake{})
	go s.Serve(l)
	t.Cleanup(s.Stop)
	*sockAddr = "unix://" + sock
	client = nil
	t.Cleanup(func() { *sockAddr, client = "", nil })
}

func TestGet(t *testing.T) {
	serve(t)
	if v, err := Get(context.Background(), "db/password"); err != nil || v != "v" {
		t.Fatalf("Get = %q, %v", v, err)
	}
	if _, err := Get(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if _, err := Get(context.Background(), "other"); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("unavailable: %v", err)
	}
}

func TestGetWithoutSocket(t *testing.T) {
	*sockAddr, client = "", nil
	if _, err := Get(context.Background(), "a"); !errors.Is(err, errNoSocket) {
		t.Fatalf("Get without a socket: %v", err)
	}
}
