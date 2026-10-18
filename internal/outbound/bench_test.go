package outbound

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/tracer"
)

func BenchmarkRoundTrip(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.MaxIdleConnsPerHost = 256
	defer base.CloseIdleConnections()

	for _, c := range []struct {
		name string
		rt   http.RoundTripper
	}{{"Base", base}, {"Traced", &Transport{Base: base}}} {
		get := func() error {
			req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
			resp, err := c.rt.RoundTrip(req)
			if err != nil {
				return err
			}
			io.Copy(io.Discard, resp.Body)
			return resp.Body.Close()
		}
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := get(); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(c.name+"Parallel", func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if err := get(); err != nil {
						b.Error(err)
						return
					}
				}
			})
		})
	}
}

type stub struct{}

func (stub) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func BenchmarkWrapper(b *testing.B) {
	rt := &Transport{Base: stub{}}
	req, _ := http.NewRequest(http.MethodGet, "http://api.example.com/", nil)
	b.Run("Serial", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			rt.RoundTrip(req)
		}
	})
	b.Run("Parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				rt.RoundTrip(req)
			}
		})
	})
}

var errNoTrace = errors.New("no trace headers")

type traced struct{}

func (traced) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header["X-Trace-Id"] == nil {
		return nil, errNoTrace
	}
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func BenchmarkProxyConnectHeader(b *testing.B) {
	opentracing.SetGlobalTracer(tracer.New(io.Discard))
	b.Cleanup(func() { opentracing.SetGlobalTracer(opentracing.NoopTracer{}) })
	_, ctx := opentracing.StartSpanFromContext(context.Background(), "parent")
	proxyURL, _ := url.Parse("http://127.0.0.1:3128")
	base := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	base.RegisterProtocol("http", traced{})
	rt := &Transport{Base: base}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://api.example.com/", nil)

	for _, c := range []struct {
		name string
		f    func() error
	}{
		{"Connect", func() error {
			if h, _ := ProxyConnectHeader(ctx, nil, ""); h == nil {
				return errNoTrace
			}
			return nil
		}},
		{"PlainHTTP", func() error {
			_, err := rt.RoundTrip(req)
			return err
		}},
	} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := c.f(); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(c.name+"Parallel", func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if err := c.f(); err != nil {
						b.Error(err)
						return
					}
				}
			})
		})
	}
}
