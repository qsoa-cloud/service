package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/tracer"

	"gopkg.qsoa.cloud/service"
)

type myTestServer struct {
	*httptest.Server
}

func (s myTestServer) ListenAndServe() error {
	s.Start()
	return nil
}
func TestTracing(t *testing.T) {
	service.Run()

	Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, _ := opentracing.StartSpanFromContext(r.Context(), "test")
		defer span.Finish()

		_, _ = w.Write([]byte(strconv.FormatUint(span.(*tracer.Span).Ctx.TraceID, 16)))
	}))

	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	server = myTestServer{s}

	req, err := http.NewRequest(http.MethodGet, s.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-TRACE-ID", strconv.FormatUint(0xAABBCCDDEEFF, 16))
	req.Header.Set("X-SPAN-ID", strconv.FormatUint(0x001122334455, 16))

	resp, err := s.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "aabbccddeeff" {
		t.Fatalf("Invalid TraceId, expected 'aabbccddeeff', got '%s'", data)
	}
}
