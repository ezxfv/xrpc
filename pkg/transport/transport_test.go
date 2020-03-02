package transport_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"x.io/xrpc/pkg/net"

	"golang.org/x/net/http2"
)

var (
	addr     = "localhost:13142"
	config   *tls.Config
	certFile = "../../testdata/server.crt"
	keyFile  = "../../testdata/server.key"
	protocol = net.UDP
)

func init() {
	// Create a pool with the server certificate since it is not signed
	// by a known CA
	cwd, _ := os.Getwd()
	println(cwd)
	caCert, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Fatalf("reading server certificate: %s", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	c, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("reading tls config: %s", err)
	}
	// Create TLS configuration with the certificate of the server
	config = &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{c},
	}
}

func TestHttp2Server(t *testing.T) {
	quit := make(chan struct{})
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("hello"))
			quit <- struct{}{}
		}),
		TLSConfig: config,
	}
	if err := http2.ConfigureServer(server, nil); err != nil {
		log.Fatal(err.Error())
		return
	}

	lis, err := net.Listen(nil, protocol, addr)
	if err != nil {
		log.Fatal(err.Error())
	}
	lis = tls.NewListener(lis, config)
	go server.Serve(lis)
	<-quit
	time.Sleep(time.Microsecond)
}

func TestHttp2Client(t *testing.T) {
	transport := http2.Transport{
		TLSClientConfig: config,
		DialTLS: func(network, adr string, cfg *tls.Config) (net.Conn, error) {
			conn, err := net.Dial(nil, protocol, addr)
			if err != nil {
				return nil, err
			}
			return net.WrapTLSClient(conn, cfg, net.DefaultTimeout)
		},
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Get("https://" + addr)
	if err != nil {
		println(err.Error())
		return
	}
	if resp.Body != nil {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err.Error())
		}
		println(string(data))
	}
}
