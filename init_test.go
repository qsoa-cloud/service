package service_test

import (
	"testing"

	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	"gopkg.qsoa.cloud/grpccodec"
	"gopkg.qsoa.cloud/grpccodec/inittest"
	_ "gopkg.qsoa.cloud/service/qdfs"
	_ "gopkg.qsoa.cloud/service/qemail"
	_ "gopkg.qsoa.cloud/service/qmysql"
)

func TestInit(t *testing.T) {
	inittest.Check(t)
	if encoding.GetCodecV2(proto.Name) == grpccodec.Codec {
		t.Fatal("the service library must not replace the global proto codec")
	}
}
