//go:generate protoc -I internal/dfspb --go_out=internal/dfspb --go-grpc_out=internal/dfspb internal/dfspb/dfs.proto
package qdfs

import (
	"context"
	"io"
	"os"

	"google.golang.org/grpc/status"

	"gopkg.qsoa.cloud/service/qdfs/internal/dfspb"
)

type Dfs struct {
	bucket string
	client dfspb.DfsClient
}

func (f *Dfs) Mkdir(ctx context.Context, name string) error {
	_, err := f.client.MkDir(ctx, &dfspb.MkDirReq{
		Bucket:   f.bucket,
		Filename: name,
	})

	return FileErrorFromGrpc(err)
}

func (f *Dfs) OpenFile(ctx context.Context, name string, flag int) (*DfsFile, error) {
	fileClient, err := f.client.File(ctx)
	if err != nil {
		return nil, FileErrorFromGrpc(err)
	}

	if err := fileClient.Send(&dfspb.FileReq{
		Msg: &dfspb.FileReq_Open_{Open: &dfspb.FileReq_Open{
			Bucket:   f.bucket,
			Filename: name,
			Flag:     int64(flag),
		}}}); err != nil {
		return nil, FileErrorFromGrpc(err)
	}

	resp, err := fileClient.Recv()
	if err != nil {
		return nil, FileErrorFromGrpc(err)
	}
	if err := resp.GetOpen().Error; err != nil {
		return nil, FileErrorFromPb(err)
	}

	return &DfsFile{fileClient}, nil
}

func (f *Dfs) RemoveAll(ctx context.Context, name string) error {
	_, err := f.client.RemoveAll(ctx, &dfspb.RemoveAllReq{
		Bucket:   f.bucket,
		Filename: name,
	})

	return FileErrorFromGrpc(err)
}

func (f *Dfs) Rename(ctx context.Context, oldName, newName string) error {
	_, err := f.client.Rename(ctx, &dfspb.RenameReq{
		Bucket:  f.bucket,
		OldName: oldName,
		NewName: newName,
	})

	return FileErrorFromGrpc(err)
}

func (f *Dfs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	resp, err := f.client.Stat(ctx, &dfspb.StatReq{
		Bucket:   f.bucket,
		Filename: name,
	})
	if err != nil {
		return nil, FileErrorFromGrpc(err)
	}

	return &FileInfo{resp}, nil
}

func FileErrorFromGrpc(err error) error {
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case 256:
			return os.ErrNotExist
		case 257:
			return os.ErrExist
		case 258:
			return io.EOF
		case 259:
			return io.ErrUnexpectedEOF
		}
	}

	return err
}
