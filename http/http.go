package http

import (
	"log"
	"net/http"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/service"
)

var (
	server      httpListenAndServe
	handlersMux *http.ServeMux = http.NewServeMux()
)

type httpListenAndServe interface {
	ListenAndServe() error
}

func Handle(location string, handler http.Handler) {
	handlersMux.Handle(location, handler)
}

func Run() {
	service.Run()

	if server == nil {
		server = &http.Server{Addr: service.GetListenAddr(), Handler: http.HandlerFunc(handler)}
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	parentSpanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		log.Fatal(err)
	}

	span, ctx := opentracing.StartSpanFromContext(r.Context(), r.Method+" "+r.URL.Path, opentracing.ChildOf(parentSpanCtx))
	defer span.Finish()

	span.SetTag("http", nil)
	span.SetTag("method", r.Method)
	span.SetTag("url", r.URL.String())

	handlersMux.ServeHTTP(w, r.WithContext(ctx))
}
