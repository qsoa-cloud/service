package http

import (
	"log"
	"net/http"

	"github.com/opentracing/opentracing-go"
	tlog "github.com/opentracing/opentracing-go/log"

	"gopkg.qsoa.cloud/service"
)

var handlersMux *http.ServeMux = http.NewServeMux()

func Handle(location string, handler http.Handler) {
	handlersMux.Handle(location, handler)
}

func Run() {
	service.Run()

	if err := http.ListenAndServe(service.GetListenAddr(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, ctx := opentracing.StartSpanFromContext(r.Context(), r.Method+" "+r.URL.Path)
		defer span.Finish()

		span.SetTag("http", nil)

		span.LogFields(tlog.String("IP", r.RemoteAddr))
		if ref := r.Referer(); ref != "" {
			tlog.String("Referrer", ref)
		}

		if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(w.Header())); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logResponse := response{w, http.StatusOK, 0}
		handlersMux.ServeHTTP(&logResponse, r.WithContext(ctx))

		span.LogFields(
			tlog.Int("Response body size", logResponse.bodySize),
			tlog.Int("Response code", logResponse.statusCode),
		)
	})); err != nil {
		log.Fatal(err)
	}
}
