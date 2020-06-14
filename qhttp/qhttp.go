package qhttp

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/service"
)

type qHttp struct {
	mux *http.ServeMux
}

var (
	httpService = &qHttp{
		mux: http.NewServeMux(),
	}
)

func init() {
	service.RegisterService(httpService)
}

func (*qHttp) GetName() string {
	return "http"
}

func (s *qHttp) Serve(l net.Listener, wg *sync.WaitGroup) {
	srv := &http.Server{
		Handler: s,
	}

	// Graceful shutdown on Interrupt signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	go func() {
		<-sigC
		ctx, _ := context.WithTimeout(context.Background(), time.Minute)
		_ = srv.Shutdown(ctx)

		wg.Done()
	}()

	// Run server
	if err := srv.Serve(l); err != nil {
		log.Fatalf("Cannot serve HTTP server: %v", err)
	}
}

func (s *qHttp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parentSpanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		// No trace
		s.mux.ServeHTTP(w, r)
	}

	span, ctx := opentracing.StartSpanFromContext(r.Context(), r.Method+" "+r.URL.Path, opentracing.ChildOf(parentSpanCtx))
	defer span.Finish()

	span.SetTag("http", nil)

	s.mux.ServeHTTP(w, r.WithContext(ctx))
}

func Handle(location string, handler http.Handler) {
	httpService.mux.Handle(location, handler)
}
