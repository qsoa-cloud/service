package service

import (
	"net/http"

	"gopkg.qsoa.cloud/service/internal/outbound"
)

func init() {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		t.GetProxyConnectHeader = outbound.ProxyConnectHeader
	}
	if http.DefaultClient.Transport == nil {
		http.DefaultClient.Transport = &outbound.Transport{Base: http.DefaultTransport}
	}
}
