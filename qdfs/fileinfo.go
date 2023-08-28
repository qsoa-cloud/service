package qdfs

import (
	"io/fs"
	"time"

	"gopkg.qsoa.cloud/service/qdfs/internal/dfspb"
)

type FileInfo struct {
	*dfspb.StatResp
}

func (f *FileInfo) Type() fs.FileMode {
	return f.Mode()
}

func (f *FileInfo) Info() (fs.FileInfo, error) {
	return f, nil
}

func (f *FileInfo) Name() string {
	return f.StatResp.Name
}

func (f *FileInfo) Size() int64 {
	return f.StatResp.Size
}

func (f *FileInfo) Mode() fs.FileMode {
	return 0644
}

func (f *FileInfo) ModTime() time.Time {
	return time.Unix(f.StatResp.ModTime, 0)
}

func (f *FileInfo) IsDir() bool {
	return f.StatResp.IsDir
}

func (f *FileInfo) Sys() any {
	return nil
}
