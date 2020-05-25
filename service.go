package service

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.qsoa.cloud/tracer"
)

var (
	version *bool = flag.Bool("q_libversion", false, "print service library version")

	project = flag.String("q_project", "", "project id")
	env     = flag.String("q_env", "", "env id")
	service = flag.String("q_service", "", "service id")

	tracerFile  = flag.String("q_tracer_file", "", "tracer file")
	metricsAddr = flag.String("q_metrics_addr", "", "http metrics addr")

	netAddrRe = regexp.MustCompile(`^(?:([\w]+)://)?(.+$)`)
)

func Run() {
	var httpAddr *string
	if hasHttpHandlers {
		httpAddr = flag.String("q_http_addr", "localhost:8080", "network type for HTTP server")
	}

	flag.Parse()

	if *version {
		fmt.Println("1.0")
		os.Exit(0)
	}

	var tracerW io.Writer = os.Stderr
	tracerFile := *tracerFile
	if tracerFile != "" {
		f, err := os.OpenFile(tracerFile, os.O_WRONLY|os.O_APPEND, 0700)
		if err != nil {
			log.Fatalf("Cannot open tracer file: %s", err.Error())
		}

		tracerW = f
	}

	opentracing.SetGlobalTracer(tracer.New(tracerW))

	if *metricsAddr != "" {
		go func() {
			if err := http.Serve(qListen(*metricsAddr), promhttp.Handler()); err != nil {
				log.Fatalf("Cannot serv metrics server: %v", err)
			}
		}()
	}

	wg := &sync.WaitGroup{}

	if hasHttpHandlers && httpAddr != nil {
		sNet, sAddr := splitNetAddr(*httpAddr)
		if sNet != "unix" {
			log.Printf("HTTP server listens on http://%s", sAddr)
		}
		serveHttp(qListen(*httpAddr), wg)
	}

	wg.Wait()
}

func GetProject() string {
	return *project
}

func GetEnv() string {
	return *env
}

func GetService() string {
	return *service
}

func splitNetAddr(addr string) (string, string) {
	addrParts := netAddrRe.FindStringSubmatch(addr)
	if len(addrParts) != 3 {
		log.Fatalf("Invalid address '%s'", addr)
	}
	if addrParts[1] == "" {
		addrParts[1] = "tcp"
	}

	return addrParts[1], addrParts[2]
}

func qListen(addr string) net.Listener {
	sNet, sAddr := splitNetAddr(addr)
	l, err := net.Listen(sNet, sAddr)
	if err != nil {
		log.Fatalf("Cannot listen %s: %v", addr, err)
	}

	if strings.HasPrefix(sNet, "unix") {
		if err := os.Chmod(sAddr, os.ModeSocket|0660); err != nil {
			log.Fatalf("Cannot change socket %s permissions: %v", addr, err)
		}
	}

	return l
}
