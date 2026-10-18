package outbound

import (
	"context"
	"maps"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"gopkg.qsoa.cloud/tracer"
)

type Transport struct {
	Base http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	span, ctx := opentracing.StartSpanFromContext(req.Context(), "HTTP "+req.Method+" "+req.URL.Host)
	defer span.Finish()
	span.SetTag("http.method", req.Method)
	span.SetTag("http.host", req.URL.Host)
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	req = req.WithContext(ctx)
	if req.URL.Scheme == "http" && viaProxy(base, req) {
		if h, _ := ProxyConnectHeader(ctx, nil, ""); h != nil {
			hdr := make(http.Header, len(req.Header)+len(h))
			maps.Copy(hdr, req.Header)
			maps.Copy(hdr, h)
			req.Header = hdr
		}
	}
	resp, err := base.RoundTrip(req)
	if err != nil {
		span.SetTag("error", nil)
		span.LogFields(log.Error(err))
		return nil, err
	}
	span.SetTag("http.status_code", resp.StatusCode)
	return resp, nil
}

func viaProxy(base http.RoundTripper, req *http.Request) bool {
	bt, ok := base.(*http.Transport)
	if !ok || bt.Proxy == nil {
		return false
	}
	u, err := bt.Proxy(req)
	return err == nil && u != nil && (u.Scheme == "http" || u.Scheme == "https")
}

// ProxyConnectHeader returns the trace headers of the span in ctx.
func ProxyConnectHeader(ctx context.Context, _ *url.URL, _ string) (http.Header, error) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil, nil
	}
	if _, ok := span.Context().(*tracer.SpanContext); !ok {
		return nil, nil
	}
	h := http.Header{}
	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(h)); err != nil {
		return nil, nil
	}
	return h, nil
}
