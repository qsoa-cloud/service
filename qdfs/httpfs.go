package qdfs

import (
	"context"
	"net/http"
	"os"
)

var _ http.FileSystem = &HttpFs{}

type HttpFs struct {
	*Dfs
}

func (f *HttpFs) Open(name string) (http.File, error) {
	return f.OpenFile(context.Background(), name, os.O_RDONLY)
}

func (f *Dfs) ToHttpFs() *HttpFs {
	return &HttpFs{f}
}
