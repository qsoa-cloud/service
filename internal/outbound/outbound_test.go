package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/tracer"
)

type span struct {
	O string `json:"o"`
	C struct {
		T uint64 `json:"t"`
		S uint64 `json:"s"`
		P uint64 `json:"p"`
	} `json:"c"`
	T map[string]any `json:"t"`
}

func capture(t *testing.T) *bytes.Buffer {
	buf := &bytes.Buffer{}
	opentracing.SetGlobalTracer(tracer.New(buf))
	t.Cleanup(func() { opentracing.SetGlobalTracer(opentracing.NoopTracer{}) })
	return buf
}

func spanOf(t *testing.T, buf *bytes.Buffer, op string) (span, string) {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var s span
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			t.Fatal(err)
		}
		if s.O == op {
			return s, line
		}
	}
	t.Fatalf("no span %q in %s", op, buf)
	return span{}, ""
}

func hex(id uint64) string { return strconv.FormatUint(id, 16) }

func parentCtx() (*tracer.SpanContext, context.Context) {
	sp, ctx := opentracing.StartSpanFromContext(context.Background(), "parent")
	return sp.Context().(*tracer.SpanContext), ctx
}

func headers(srv chan<- http.Header) http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) { srv <- r.Header }
}

func roundTrip(t *testing.T, rt http.RoundTripper, req *http.Request) {
	t.Helper()
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestRoundTrip(t *testing.T) {
	buf := capture(t)
	got := make(chan http.Header, 1)
	srv := httptest.NewServer(headers(got))
	defer srv.Close()

	parent, ctx := parentCtx()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/secret/path?token=hunter2", nil)
	roundTrip(t, &Transport{Base: http.DefaultTransport}, req)

	host := strings.TrimPrefix(srv.URL, "http://")
	s, line := spanOf(t, buf, "HTTP GET "+host)
	if s.C.T != parent.TraceID || s.C.P != parent.SpanID {
		t.Errorf("span %+v is not a child of %+v", s.C, parent)
	}
	want := map[string]any{"http.method": "GET", "http.host": host, "http.status_code": float64(200)}
	if !reflect.DeepEqual(s.T, want) {
		t.Errorf("tags %v, want %v", s.T, want)
	}
	if strings.Contains(line, "secret") || strings.Contains(line, "hunter2") {
		t.Errorf("path or query in span: %s", line)
	}
	if h := <-got; h.Get("X-Trace-Id") != "" || h.Get("X-Span-Id") != "" {
		t.Errorf("trace headers without a proxy: %v", h)
	}
}

func TestRoundTripError(t *testing.T) {
	buf := capture(t)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()

	req, _ := http.NewRequest(http.MethodGet, "http://"+addr+"/", nil)
	if _, err := (&Transport{Base: http.DefaultTransport}).RoundTrip(req); err == nil {
		t.Fatal("no dial error")
	}
	s, _ := spanOf(t, buf, "HTTP GET "+addr)
	if _, ok := s.T["error"]; !ok {
		t.Errorf("no error tag: %v", s.T)
	}
}

func TestPlainHTTPViaProxy(t *testing.T) {
	buf := capture(t)
	got := make(chan http.Header, 1)
	proxy := httptest.NewServer(headers(got))
	defer proxy.Close()
	proxyURL, _ := url.Parse(proxy.URL)
	base := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	defer base.CloseIdleConnections()

	_, ctx := parentCtx()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://api.example.com/secret?token=hunter2", nil)
	roundTrip(t, &Transport{Base: base}, req)

	s, _ := spanOf(t, buf, "HTTP GET api.example.com")
	if h := <-got; h.Get("X-Trace-Id") != hex(s.C.T) || h.Get("X-Span-Id") != hex(s.C.S) {
		t.Errorf("proxy got %v, want trace %x span %x", h, s.C.T, s.C.S)
	}
	if len(req.Header) != 0 {
		t.Errorf("request modified: %v", req.Header)
	}
}

func TestHTTPSViaProxy(t *testing.T) {
	buf := capture(t)
	got := make(chan http.Header, 1)
	target := httptest.NewTLSServer(headers(got))
	defer target.Close()
	connect := make(chan http.Header, 1)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "CONNECT only", http.StatusBadRequest)
			return
		}
		connect <- r.Header
		dst, err := net.Dial("tcp", r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		c, rw, err := w.(http.Hijacker).Hijack()
		if err != nil {
			dst.Close()
			return
		}
		defer c.Close()
		io.WriteString(c, "HTTP/1.1 200 OK\r\n\r\n")
		go func() {
			io.Copy(dst, rw)
			dst.Close()
		}()
		io.Copy(c, dst)
	}))
	defer proxy.Close()
	proxyURL, _ := url.Parse(proxy.URL)
	base := &http.Transport{
		Proxy:                 http.ProxyURL(proxyURL),
		TLSClientConfig:       target.Client().Transport.(*http.Transport).TLSClientConfig,
		GetProxyConnectHeader: ProxyConnectHeader,
	}
	defer base.CloseIdleConnections()

	_, ctx := parentCtx()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target.URL+"/", nil)
	roundTrip(t, &Transport{Base: base}, req)

	if h := <-got; h.Get("X-Trace-Id") != "" || h.Get("X-Span-Id") != "" {
		t.Errorf("trace headers inside the tunnel: %v", h)
	}
	s, _ := spanOf(t, buf, "HTTP GET "+strings.TrimPrefix(target.URL, "https://"))
	if h := <-connect; h.Get("X-Trace-Id") != hex(s.C.T) || h.Get("X-Span-Id") != hex(s.C.S) {
		t.Errorf("CONNECT got %v, want trace %x span %x", h, s.C.T, s.C.S)
	}
}

func TestViaProxy(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://api.example.com/", nil)
	for proxy, want := range map[string]bool{
		"http://127.0.0.1:3128":   true,
		"https://127.0.0.1:3128":  true,
		"socks5://127.0.0.1:1080": false,
	} {
		u, _ := url.Parse(proxy)
		if got := (&Transport{Base: &http.Transport{Proxy: http.ProxyURL(u)}}).viaProxy(req); got != want {
			t.Errorf("%s: %v, want %v", proxy, got, want)
		}
	}
	if (&Transport{Base: &http.Transport{}}).viaProxy(req) {
		t.Error("no proxy")
	}
	if (&Transport{Base: http.NewFileTransport(http.Dir("."))}).viaProxy(req) {
		t.Error("not an *http.Transport")
	}
}

func TestProxyConnectHeader(t *testing.T) {
	capture(t)
	sc, ctx := parentCtx()
	h, err := ProxyConnectHeader(ctx, nil, "")
	if err != nil || len(h) != 2 || h.Get("X-Trace-Id") != hex(sc.TraceID) || h.Get("X-Span-Id") != hex(sc.SpanID) {
		t.Errorf("got %v, %v; want trace %x span %x", h, err, sc.TraceID, sc.SpanID)
	}
	if h, err := ProxyConnectHeader(context.Background(), nil, ""); h != nil || err != nil {
		t.Errorf("no span: %v, %v", h, err)
	}
}

func TestProxyConnectHeaderNoop(t *testing.T) {
	opentracing.SetGlobalTracer(opentracing.NoopTracer{})
	_, ctx := opentracing.StartSpanFromContext(context.Background(), "noop")
	if h, err := ProxyConnectHeader(ctx, nil, ""); h != nil || err != nil {
		t.Errorf("noop tracer: %v, %v", h, err)
	}
	capture(t)
	if h, err := ProxyConnectHeader(ctx, nil, ""); h != nil || err != nil {
		t.Errorf("noop span: %v, %v", h, err)
	}
}
