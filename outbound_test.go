package service

import (
	"net/http"
	"reflect"
	"testing"

	"gopkg.qsoa.cloud/service/internal/outbound"
)

func TestOutbound(t *testing.T) {
	dt, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Fatalf("DefaultTransport is %T", http.DefaultTransport)
	}
	if reflect.ValueOf(dt.GetProxyConnectHeader).Pointer() != reflect.ValueOf(outbound.ProxyConnectHeader).Pointer() {
		t.Error("GetProxyConnectHeader is not outbound.ProxyConnectHeader")
	}
	if ct, ok := http.DefaultClient.Transport.(*outbound.Transport); !ok || ct.Base != nil {
		t.Errorf("DefaultClient.Transport is %#v", http.DefaultClient.Transport)
	}
	if c := dt.Clone(); c == nil || c.GetProxyConnectHeader == nil {
		t.Error("Clone lost GetProxyConnectHeader")
	}
}

type stubTransport struct{ req *http.Request }

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.req = req
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}, nil
}

func TestDefaultTransportSwap(t *testing.T) {
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })
	stub := &stubTransport{}
	http.DefaultTransport = stub

	resp, err := http.DefaultClient.Get("http://127.0.0.1:1/")
	if err != nil {
		t.Fatalf("DefaultClient bypassed the swapped DefaultTransport: %v", err)
	}
	resp.Body.Close()
	if stub.req == nil || stub.req.URL.Host != "127.0.0.1:1" {
		t.Errorf("stub got %v", stub.req)
	}
}
