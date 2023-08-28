package qdfs

import (
	"context"
	"io/fs"
	"os"
)

var _ fs.FS = &FsFs{}

type FsFs struct {
	*Dfs
}

func (f *FsFs) Open(name string) (fs.File, error) {
	return f.OpenFile(context.Background(), name, os.O_RDONLY)
}

func (f *Dfs) ToFsFs() *FsFs {
	return &FsFs{f}
}
