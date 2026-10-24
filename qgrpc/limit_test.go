package qgrpc

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func serveLimit(t *testing.T, s *grpc.Server) string {
	t.Helper()
	handler := func(size func(*wrapperspb.BytesValue) int) grpc.MethodHandler {
		return func(_ any, _ context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
			var m wrapperspb.BytesValue
			if err := dec(&m); err != nil {
				return nil, err
			}
			return wrapperspb.Bytes(make([]byte, size(&m))), nil
		}
	}
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "limit.Limit",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: handler(func(m *wrapperspb.BytesValue) int { return len(m.Value) })},
			{MethodName: "Fill", Handler: handler(func(*wrapperspb.BytesValue) int { return 17 << 20 })},
			{MethodName: "Sink", Handler: handler(func(*wrapperspb.BytesValue) int { return 0 })},
		},
	}, struct{}{})
	sock := filepath.Join(t.TempDir(), "grpc.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	go s.Serve(l)
	t.Cleanup(s.Stop)
	return "unix://" + sock
}

func checkLimit(t *testing.T, conn *grpc.ClientConn) {
	t.Helper()
	var out wrapperspb.BytesValue
	if err := conn.Invoke(context.Background(), "/limit.Limit/Echo", wrapperspb.Bytes(make([]byte, 15<<20)), &out); err != nil || len(out.Value) != 15<<20 {
		t.Fatalf("15 MiB: %d, %v", len(out.Value), err)
	}
	if err := conn.Invoke(context.Background(), "/limit.Limit/Sink", wrapperspb.Bytes(make([]byte, 17<<20)), &out); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("17 MiB request: %v", err)
	}
	if err := conn.Invoke(context.Background(), "/limit.Limit/Fill", &wrapperspb.BytesValue{}, &out); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("17 MiB response: %v", err)
	}
}

func TestServerMessageLimit(t *testing.T) {
	conn, err := grpc.NewClient(serveLimit(t, newServer()), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(64<<20)))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	checkLimit(t, conn)
}

func TestClientMessageLimit(t *testing.T) {
	conn, err := Dial(serveLimit(t, grpc.NewServer(grpc.MaxRecvMsgSize(64<<20), grpc.MaxSendMsgSize(64<<20))))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	checkLimit(t, conn)
}
