package qhttp

import (
	"net/http"

	"gopkg.qsoa.cloud/service/internal/outbound"
)

// Transport wraps the current http.DefaultTransport with a span per request.
func Transport() http.RoundTripper {
	return &outbound.Transport{Base: http.DefaultTransport}
}
