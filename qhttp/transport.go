package qhttp

import (
	"net/http"

	"gopkg.qsoa.cloud/service/internal/outbound"
)

// Transport returns http.DefaultTransport with a span per request.
func Transport() http.RoundTripper {
	return &outbound.Transport{}
}
