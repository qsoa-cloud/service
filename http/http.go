package http

import (
	"log"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/service"
)

var (
	handlersMux = http.NewServeMux()
)

type httpResponseWriter struct {
	w      http.ResponseWriter
	status int
}

func (w *httpResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *httpResponseWriter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *httpResponseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
	w.status = code
}

func Handle(location string, handler http.Handler) {
	handlersMux.Handle(location, handler)
}

func Run() {
	service.Run()

	server := &http.Server{
		Addr:      service.GetListenAddr(),
		Handler:   http.HandlerFunc(handler),
		TLSConfig: service.GetServerTlsConfig(),
	}

	if service.GetServerCert() != "" || service.GetServerPrivKey() != "" {
		if err := server.ListenAndServeTLS(service.GetServerCert(), service.GetServerPrivKey()); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	parentSpanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		log.Fatal(err)
	}

	span, ctx := opentracing.StartSpanFromContext(r.Context(), r.Method+" "+r.URL.Path, opentracing.ChildOf(parentSpanCtx))
	defer span.Finish()

	span.SetTag("http", nil)
	span.SetTag("method", r.Method)
	span.SetTag("url", r.URL.String())

	service.CountRequest(r.URL.Path)

	wrappedW := &httpResponseWriter{w: w}
	handlersMux.ServeHTTP(wrappedW, r.WithContext(ctx))

	if wrappedW.status == 0 {
		wrappedW.status = http.StatusOK
	}

	service.CountResponse(r.URL.Path, wrappedW.status)
	service.CountResponseTime(r.URL.Path, time.Now().Sub(start))
}
