package main

import (
	"net/http"

	"gopkg.qsoa.cloud/service"
)

func main() {
	service.HandleHttp("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello world"))
	}))

	service.Run()

}
