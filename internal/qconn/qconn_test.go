package qconn

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	"google.golang.org/grpc"

	"gopkg.qsoa.cloud/grpccodec"
)

type probe struct{}

func (probe) Size() int                                { return 0 }
func (probe) MarshalToSizedBuffer([]byte) (int, error) { return 0, nil }
func (*probe) Unmarshal([]byte) error                  { return nil }
func (*probe) Reset()                                  {}

func TestDialUsesCodec(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "s")
	lis, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer(grpc.ForceServerCodecV2(grpccodec.Codec), grpc.UnknownServiceHandler(func(_ any, st grpc.ServerStream) error {
		var m probe
		if err := st.RecvMsg(&m); err != nil {
			return err
		}
		return st.SendMsg(&m)
	}))
	go s.Serve(lis)
	defer s.Stop()
	conn, err := Dial("unix://" + sock)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.Invoke(context.Background(), "/probe.Probe/Echo", &probe{}, &probe{}); err != nil {
		t.Fatal(err)
	}
}
