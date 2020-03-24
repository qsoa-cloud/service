package service

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCountRequest(t *testing.T) {
	CountRequest("/test/")
	CountRequest("/test2/")

	g, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range g {
		fmt.Println(m.String())
	}
}
