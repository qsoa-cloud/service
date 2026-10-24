package qgrpc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGatewayClientCert(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "client"}, NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	esc := url.QueryEscape(string(certPEM))

	for _, c := range []struct {
		name string
		md   metadata.MD
		ok   bool
	}{
		{"gateway", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {esc}}, true},
		{"service", metadata.MD{"x-qcloud-via": {"service"}, "x-client-cert": {esc}}, false},
		{"no via", metadata.MD{"x-client-cert": {esc}}, false},
		{"two via", metadata.MD{"x-qcloud-via": {"gateway", "gateway"}, "x-client-cert": {esc}}, false},
		{"no cert", metadata.MD{"x-qcloud-via": {"gateway"}}, false},
		{"two certs", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {esc, esc}}, false},
		{"bad escape", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {esc + "%zz"}}, false},
		{"garbage", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {"garbage"}}, false},
		{"not a certificate", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {url.QueryEscape(string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))}}, false},
		{"bad der", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {url.QueryEscape(string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("garbage")})))}}, false},
		{"two pems", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {url.QueryEscape(string(certPEM) + string(certPEM))}}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			cert, ok := GatewayClientCert(metadata.NewIncomingContext(context.Background(), c.md))
			if ok != c.ok || (ok && !bytes.Equal(cert.Raw, der)) || (!ok && cert != nil) {
				t.Fatalf("GatewayClientCert: cert %t, ok %t", cert != nil, ok)
			}
		})
	}
	if cert, ok := GatewayClientCert(context.Background()); ok || cert != nil {
		t.Fatalf("no metadata: cert %t, ok %t", cert != nil, ok)
	}
}

func TestServerMessageLimit(t *testing.T) {
	handler := func(size func(*wrapperspb.BytesValue) int) grpc.MethodHandler {
		return func(_ any, _ context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
			var m wrapperspb.BytesValue
			if err := dec(&m); err != nil {
				return nil, err
			}
			return wrapperspb.Bytes(make([]byte, size(&m))), nil
		}
	}
	s := GetServer()
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "limit.Limit",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: handler(func(m *wrapperspb.BytesValue) int { return len(m.Value) })},
			{MethodName: "Fill", Handler: handler(func(*wrapperspb.BytesValue) int { return 17 << 20 })},
		},
	}, struct{}{})
	sock := filepath.Join(t.TempDir(), "grpc.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	go s.Serve(l)
	t.Cleanup(s.Stop)
	conn, err := grpc.NewClient("unix://"+sock, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(64<<20)))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var out wrapperspb.BytesValue
	if err := conn.Invoke(context.Background(), "/limit.Limit/Echo", wrapperspb.Bytes(make([]byte, 15<<20)), &out); err != nil || len(out.Value) != 15<<20 {
		t.Fatalf("15 MiB: %d, %v", len(out.Value), err)
	}
	if err := conn.Invoke(context.Background(), "/limit.Limit/Echo", wrapperspb.Bytes(make([]byte, 17<<20)), &out); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("17 MiB request: %v", err)
	}
	if err := conn.Invoke(context.Background(), "/limit.Limit/Fill", &wrapperspb.BytesValue{}, &out); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("17 MiB response: %v", err)
	}
}
