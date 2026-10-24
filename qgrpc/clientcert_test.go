package qgrpc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"
)

func TestGatewayClientCert(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "client"}, NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	esc := url.QueryEscape(string(certPEM))

	for _, c := range []struct {
		name string
		md   metadata.MD
		ok   bool
	}{
		{"gateway", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {esc}}, true},
		{"service", metadata.MD{"x-qcloud-via": {"service"}, "x-client-cert": {esc}}, false},
		{"no via", metadata.MD{"x-client-cert": {esc}}, false},
		{"two via", metadata.MD{"x-qcloud-via": {"gateway", "gateway"}, "x-client-cert": {esc}}, false},
		{"no cert", metadata.MD{"x-qcloud-via": {"gateway"}}, false},
		{"two certs", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {esc, esc}}, false},
		{"bad escape", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {esc + "%zz"}}, false},
		{"garbage", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {"garbage"}}, false},
		{"not a certificate", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {url.QueryEscape(string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))}}, false},
		{"bad der", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {url.QueryEscape(string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("garbage")})))}}, false},
		{"two pems", metadata.MD{"x-qcloud-via": {"gateway"}, "x-client-cert": {url.QueryEscape(string(certPEM) + string(certPEM))}}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			cert, ok := GatewayClientCert(metadata.NewIncomingContext(context.Background(), c.md))
			if ok != c.ok || (ok && !bytes.Equal(cert.Raw, der)) || (!ok && cert != nil) {
				t.Fatalf("GatewayClientCert: cert %t, ok %t", cert != nil, ok)
			}
		})
	}
	if cert, ok := GatewayClientCert(context.Background()); ok || cert != nil {
		t.Fatalf("no metadata: cert %t, ok %t", cert != nil, ok)
	}
}
