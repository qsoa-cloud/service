package qhttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/tracer"
)

func TestTransportAsDefault(t *testing.T) {
	buf := &bytes.Buffer{}
	opentracing.SetGlobalTracer(tracer.New(buf))
	t.Cleanup(func() { opentracing.SetGlobalTracer(opentracing.NoopTracer{}) })
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })
	http.DefaultTransport = Transport()

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	resp, err := (&http.Client{}).Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if n := strings.Count(buf.String(), "\n"); n != 1 {
		t.Errorf("%d spans: %s", n, buf)
	}
}
