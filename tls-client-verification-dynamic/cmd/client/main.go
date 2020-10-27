package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bananaumai/suburi-go/tls-client-verification-dynamic/internal"
)

var (
	serverCA   *x509.Certificate
	clientCert tls.Certificate
)

func init() {
	var (
		serverCAPath   string
		clientCertPath string
	)

	flag.StringVar(&serverCAPath, "serverCA", "", "Path to server CA file")
	flag.StringVar(&clientCertPath, "clientCert", "", "Path to the client certificate directory; There should be `client.crt` and `client.key`")
	flag.Parse()

	var err error

	if err = loadServerCA(serverCAPath); err != nil {
		log.Fatalf("failed to load server CA: %s", err)
	}

	if clientCert, err = internal.LoadTLSCert(clientCertPath, "client.crt", "client.key"); err != nil {
		log.Fatalf("failed to load client cert: %s", err)
	}
}

func main() {
	pool := x509.NewCertPool()
	pool.AddCert(serverCA)

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      pool,
	}
	transport := &http.Transport{TLSClientConfig: tlsCfg}
	client := &http.Client{Transport: transport}

	res, err := client.Get("https://localhost:443")
	if err != nil {
		log.Fatalf("http client error: %s", err)
	}
	defer func() { _ = res.Body.Close() }()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %s", err)
	}

	log.Printf("%s", string(data))
}

func loadServerCA(caPath string) error {
	var err error
	if serverCA, err = internal.LoadX509Cert(caPath); err != nil {
		return fmt.Errorf("failed to load cert file: %w", err)
	}

	return nil
}
