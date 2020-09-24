package service

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.qsoa.cloud/tracer"
)

type Service interface {
	GetName() string
	Serve(l net.Listener, wg *sync.WaitGroup)
}

// Flags
var (
	libVersion = flag.Bool("q_libversion", false, "print service library version")

	project = flag.String("q_project", "", "project id")
	env     = flag.String("q_env", "", "env id")
	service = flag.String("q_service", "", "service id")
	version = flag.String("q_version", "", "service version")

	tracerFile  = flag.String("q_tracer_file", "", "tracer file")
	metricsAddr = flag.String("q_metrics_addr", "", "http metrics addr")
)

// Internal data
var (
	services        = map[string]Service{}
	servicesAddr    = map[string]*string{}
	nextDefaultPort = 8080
)

func RegisterService(service Service) {
	RegisterClientService(service)

	name := service.GetName()
	servicesAddr[name] = flag.String(
		"q_"+name+"_addr",
		"127.0.0.1:"+strconv.FormatInt(int64(nextDefaultPort), 10),
		"network address for "+name+" server",
	)
	nextDefaultPort++
}

func RegisterClientService(service Service) {
	name := service.GetName()
	if _, exists := services[name]; exists {
		log.Fatalf("Service with name '%s' is already exists", name)
	}

	services[name] = service
}

func Run() {
	flag.Parse()

	if *libVersion {
		fmt.Println("1.0")
		os.Exit(0)
	}

	// Prepare tracer
	var tracerW io.Writer = os.Stderr
	tracerFile := *tracerFile
	if tracerFile != "" {
		f, err := os.OpenFile(tracerFile, os.O_WRONLY|os.O_APPEND, 0700)
		if err != nil {
			log.Fatalf("Cannot open tracer file: %s", err.Error())
		}
		defer f.Close()

		tracerW = f
	}

	opentracing.SetGlobalTracer(tracer.New(tracerW))

	// Prepare metrics socket
	if *metricsAddr != "" {
		sNet, sAddr := splitNetAddr(*metricsAddr)
		l := qListen(sNet, sAddr)
		defer l.Close()

		go func() {
			if err := http.Serve(l, promhttp.Handler()); err != nil {
				log.Fatalf("Cannot serv metrics server: %v", err)
			}
		}()
	}

	// Serve servers
	wg := &sync.WaitGroup{}
	for _, s := range services {
		sAddr, exists := servicesAddr[s.GetName()]
		var l net.Listener
		if exists {
			sNet, sAddr := splitNetAddr(*sAddr)
			if sNet != "unix" {
				log.Printf(s.GetName()+" server listens on %s", sAddr)
			}

			l = qListen(sNet, sAddr)
			//noinspection ALL
			defer l.Close()
		}

		wg.Add(1)
		go s.Serve(l, wg)
	}

	wg.Wait()
}

func GetProject() string {
	if !flag.Parsed() {
		flag.Parse()
	}

	return *project
}

func GetEnv() string {
	if !flag.Parsed() {
		flag.Parse()
	}

	return *env
}

func GetService() string {
	if !flag.Parsed() {
		flag.Parse()
	}

	return *service
}

func GetVersion() string {
	if !flag.Parsed() {
		flag.Parse()
	}

	return *service
}
