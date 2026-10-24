package qgrpc

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"net/url"

	"google.golang.org/grpc/metadata"
)

// GatewayClientCert returns the client certificate that the gateway verified and forwarded with the call.
func GatewayClientCert(ctx context.Context) (*x509.Certificate, bool) {
	md, _ := metadata.FromIncomingContext(ctx)
	via, raw := md.Get("x-qcloud-via"), md.Get("x-client-cert")
	if len(via) != 1 || via[0] != "gateway" || len(raw) != 1 {
		return nil, false
	}
	s, err := url.QueryUnescape(raw[0])
	if err != nil {
		return nil, false
	}
	b, rest := pem.Decode([]byte(s))
	if b == nil || b.Type != "CERTIFICATE" || len(rest) != 0 {
		return nil, false
	}
	c, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return nil, false
	}
	return c, true
}
