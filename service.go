package service

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/opentracing/opentracing-go"

	"gopkg.qsoa.cloud/tracer"
)

var (
	version *bool = flag.Bool("q_libversion", false, "print service library version")

	host       = flag.String("q_host", "localhost", "service host")
	port       = flag.Uint("q_port", 8080, "service port")
	project    = flag.String("q_project", "", "project id")
	env        = flag.String("q_env", "", "env id")
	service    = flag.String("q_service", "", "service id")
	tracerFile = flag.String("q_tracer_file", "", "tracer file")

	etcdAddr     = flag.String("q_discovery_addr", "", "discovery hosts")
	etcdUser     = flag.String("q_discovery_user", "", "discovery user")
	etcdPassword = flag.String("q_discovery_password", "", "discovery user password")
)

func Run() {
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

	log.Printf("Service started on %s", GetListenAddr())
}

func GetHost() string {
	return *host
}

func GetPort() uint16 {
	return uint16(*port)
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

func GetDiscovery() (string, string, string) {
	return *etcdAddr, *etcdUser, *etcdPassword
}

func GetListenAddr() string {
	return fmt.Sprintf("%s:%d", *host, *port)
}
