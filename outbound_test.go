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
	if ct, ok := http.DefaultClient.Transport.(*outbound.Transport); !ok || ct.Base != http.DefaultTransport {
		t.Errorf("DefaultClient.Transport is %#v", http.DefaultClient.Transport)
	}
	if c := dt.Clone(); c == nil || c.GetProxyConnectHeader == nil {
		t.Error("Clone lost GetProxyConnectHeader")
	}
}
