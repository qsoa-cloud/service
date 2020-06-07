package service

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
)

var (
	hasHttpHandlers = false
	httpHandlersMux = http.NewServeMux()
)

func HandleHttp(location string, handler http.Handler) {
	hasHttpHandlers = true
	httpHandlersMux.Handle(location, handler)
}

func serveHttp(l net.Listener, wg *sync.WaitGroup) {
	wg.Add(1)

	s := &http.Server{
		Handler: http.HandlerFunc(httpWrapper),
	}

	// Graceful shutdown on Interrupt signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	go func() {
		<-sigC
		ctx, _ := context.WithTimeout(context.Background(), time.Minute)
		_ = s.Shutdown(ctx)

		wg.Done()
	}()

	go func() {
		if err := s.Serve(l); err != nil {
			log.Fatalf("Cannot serve HTTP server: %v", err)
		}
	}()
}

func httpWrapper(w http.ResponseWriter, r *http.Request) {
	parentSpanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		// No trace
		httpHandlersMux.ServeHTTP(w, r)
	}

	span, ctx := opentracing.StartSpanFromContext(r.Context(), r.Method+" "+r.URL.Path, opentracing.ChildOf(parentSpanCtx))
	defer span.Finish()

	span.SetTag("http", nil)

	httpHandlersMux.ServeHTTP(w, r.WithContext(ctx))
}
