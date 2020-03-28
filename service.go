package service

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.qsoa.cloud/tracer"
)

var (
	version *bool = flag.Bool("q_libversion", false, "print service library version")

	host        = flag.String("q_host", "localhost", "service host")
	port        = flag.Uint("q_port", 8080, "service port")
	project     = flag.String("q_project", "", "project id")
	env         = flag.String("q_env", "", "env id")
	service     = flag.String("q_service", "", "service id")
	tracerFile  = flag.String("q_tracer_file", "", "tracer file")
	metricsAddr = flag.String("q_metrics_addr", "", "http metrics addr")

	ca         = flag.String("q_ca", "", "CA certificate file")
	serverCert = flag.String("q_server_cert", "", "Server certificate file")
	serverPriv = flag.String("q_server_priv", "", "Server private key file")
	clientCert = flag.String("q_client_cert", "", "Client certificate file")
	clientPriv = flag.String("q_client_priv", "", "Client private key file")

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

	if *metricsAddr != "" {
		addrParts := regexp.MustCompile(`^(?:([\w]+)://)?(.+$)`).FindStringSubmatch(*metricsAddr)
		if len(addrParts) != 3 {
			log.Fatalf("Invalid metrics address '%s'", *metricsAddr)
		}
		if addrParts[1] == "" {
			addrParts[1] = "tcp"
		}

		metricsSock, err := net.Listen(addrParts[1], addrParts[2])
		if err != nil {
			log.Fatalf("Cannot listen %s: %v", *metricsAddr, err)
		}

		go http.Serve(metricsSock, promhttp.Handler())
	}

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

func GetCa() string {
	return *ca
}

func GetServerCert() string {
	return *serverCert
}

func GetServerPrivKey() string {
	return *serverPriv
}

func GetClientCert() string {
	return *clientCert
}

func GetClientPrivKey() string {
	return *clientPriv
}

func GetServerTlsConfig() *tls.Config {
	if *ca == "" && *serverCert == "" && *serverPriv == "" {
		return nil
	}

	return &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  GetCaPool(),

		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
				return fmt.Errorf("invalid chains")
			}

			clientProject := strings.Join(verifiedChains[0][0].Subject.Organization, "|")
			if clientProject != "qgateway" && clientProject != *project {
				return fmt.Errorf("invalid project: %s", clientProject)
			}

			return nil
		},
	}
}

func GetClientTlsConfig() *tls.Config {
	if *ca == "" && *clientCert == "" && *clientPriv == "" {
		return nil
	}

	res := &tls.Config{
		RootCAs: GetCaPool(),
	}

	if *clientPriv != "" && *clientCert != "" {
		cert, err := tls.LoadX509KeyPair(*clientCert, *clientPriv)
		if err != nil {
			log.Fatalf("Cannot load client certifiate: %v", err)
		}

		res.Certificates = []tls.Certificate{cert}
	}

	return res
}

func GetCaPool() *x509.CertPool {
	if *ca == "" {
		return nil
	}

	caPem, err := ioutil.ReadFile(*ca)
	if err != nil {
		log.Fatalf("Cannot read CA file: %v", err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caPem)

	return certPool
}
