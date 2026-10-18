package outbound

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
