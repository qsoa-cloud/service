package qdfs

import (
	"errors"
	"io"
	"io/fs"
	"os"

	"gopkg.qsoa.cloud/service/qdfs/internal/dfspb"
)

type DfsFile struct {
	client dfspb.Dfs_FileClient
}

func NewFile(client dfspb.Dfs_FileClient) *DfsFile {
	return &DfsFile{
		client: client,
	}
}

func (f *DfsFile) Close() error {
	defer func() {
		// Close stream
		f.client.CloseSend()
		f.client.Recv()
	}()

	if err := f.client.Send(&dfspb.FileReq{Msg: &dfspb.FileReq_Close_{Close: &dfspb.FileReq_Close{}}}); err != nil {
		return err
	}

	resp, err := f.client.Recv()
	if err != nil {
		return err
	}
	if err := resp.GetClose().Error; err != nil {
		return FileErrorFromPb(err)
	}

	return nil
}

func (f *DfsFile) Seek(offset int64, whence int) (int64, error) {
	if err := f.client.Send(&dfspb.FileReq{Msg: &dfspb.FileReq_Seek_{Seek: &dfspb.FileReq_Seek{
		Offset: offset,
		Whence: int64(whence),
	}}}); err != nil {
		return 0, err
	}

	resp, err := f.client.Recv()
	if err != nil {
		return 0, err
	}

	msg := resp.GetSeek()
	if err := msg.Error; err != nil {
		return 0, FileErrorFromPb(err)
	}

	return msg.N, nil
}

func (f *DfsFile) Read(p []byte) (n int, err error) {
	if err := f.client.Send(&dfspb.FileReq{Msg: &dfspb.FileReq_Read_{Read: &dfspb.FileReq_Read{
		N: int64(len(p)),
	}}}); err != nil {
		return 0, err
	}

	resp, err := f.client.Recv()
	if err != nil {
		return 0, err
	}

	msg := resp.GetRead()
	if err := msg.Error; err != nil {
		return 0, FileErrorFromPb(err)
	}

	copy(p, msg.Data)

	return len(msg.Data), nil
}

func (f *DfsFile) Write(p []byte) (n int, err error) {
	if err := f.client.Send(&dfspb.FileReq{Msg: &dfspb.FileReq_Write_{Write: &dfspb.FileReq_Write{
		Data: p,
	}}}); err != nil {
		return 0, err
	}

	resp, err := f.client.Recv()
	if err != nil {
		return 0, err
	}

	msg := resp.GetWrite()
	if err := msg.Error; err != nil {
		return 0, FileErrorFromPb(err)
	}

	return int(msg.N), nil
}

func (f *DfsFile) Readdir(count int) ([]fs.FileInfo, error) {
	if err := f.client.Send(&dfspb.FileReq{Msg: &dfspb.FileReq_ReadDir_{ReadDir: &dfspb.FileReq_ReadDir{
		N: int64(count),
	}}}); err != nil {
		return nil, err
	}

	resp, err := f.client.Recv()
	if err != nil {
		return nil, err
	}

	msg := resp.GetReadDir()
	if err := msg.Error; err != nil {
		return nil, FileErrorFromPb(err)
	}

	res := make([]fs.FileInfo, len(msg.Files))
	for i, f := range msg.Files {
		res[i] = &FileInfo{f}
	}

	return res, nil
}

func (f *DfsFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if err := f.client.Send(&dfspb.FileReq{Msg: &dfspb.FileReq_ReadDir_{ReadDir: &dfspb.FileReq_ReadDir{
		N: int64(n),
	}}}); err != nil {
		return nil, err
	}

	resp, err := f.client.Recv()
	if err != nil {
		return nil, err
	}

	msg := resp.GetReadDir()
	if err := msg.Error; err != nil {
		return nil, FileErrorFromPb(err)
	}

	res := make([]fs.DirEntry, len(msg.Files))
	for i, f := range msg.Files {
		res[i] = &FileInfo{f}
	}

	return res, nil
}

func (f *DfsFile) Stat() (fs.FileInfo, error) {
	if err := f.client.Send(&dfspb.FileReq{Msg: &dfspb.FileReq_Stat_{Stat: &dfspb.FileReq_Stat{}}}); err != nil {
		return nil, err
	}

	resp, err := f.client.Recv()
	if err != nil {
		return nil, err
	}

	msg := resp.GetStat()
	if err := msg.Error; err != nil {
		return nil, FileErrorFromPb(err)
	}

	return &FileInfo{msg.File}, nil
}

func FileErrorFromPb(e *dfspb.FileResp_Error) error {
	if e == nil {
		return nil
	}

	switch e.Type {
	case dfspb.FileResp_Error_EXIST:
		return os.ErrExist
	case dfspb.FileResp_Error_NOT_EXIST:
		return os.ErrNotExist
	case dfspb.FileResp_Error_EOF:
		return io.EOF
	case dfspb.FileResp_Error_UNEXPECTED_EOF:
		return io.ErrUnexpectedEOF
	default:
		return errors.New(e.Msg)
	}
}
