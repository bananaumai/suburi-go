package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bananaumai/suburi-go/tls-client-verification-dynamic/internal"
)

type (
	clientCertRecord struct {
		rootCA      *x509.Certificate
		clientCerts map[string]*x509.Certificate
	}
)

var (
	serverCert   tls.Certificate
	clientCertDB []*clientCertRecord
)

func init() {
	var (
		serverCertPath string
		clientCertPath string
	)

	flag.StringVar(&serverCertPath, "serverCert", "", "Path to the server certificate directory; There should be `server.crt` and `server.key`")
	flag.StringVar(&clientCertPath, "clientCert", "", "Path to the directories that contain client certificates; Specify multiple paths by comma (`,`); There should be `ca.crt` and `client.crt` in each directory.")
	flag.Parse()

	var err error

	if serverCert, err = internal.LoadTLSCert(serverCertPath, "server.crt", "server.key"); err != nil {
		log.Fatalf("failed to load server cert: %s", err)
	}

	if clientCertDB, err = loadClientCertDB(clientCertPath); err != nil {
		log.Fatalf("failed to load client certificate database: %s", err)
	}
}

func main() {
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequestClientCert,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			// If the client sends the client certificate, verify it.
			// If there are no client certificate, make sure return no error
			// so that the http server can handle non-client cert request.
			for _, c := range rawCerts {
				clientCert, err := x509.ParseCertificate(c)
				if err != nil {
					return err
				}
				ca, err := findRootCAFromClientCertDB(clientCert)
				if err != nil {
					return err
				}
				pool := x509.NewCertPool()
				pool.AddCert(ca)
				verifyOpts := x509.VerifyOptions{
					Roots:         pool,
					Intermediates: x509.NewCertPool(),
				}
				_, err = clientCert.Verify(verifyOpts)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		log.Printf("hello")
		_, _ = w.Write([]byte("hello"))
	})

	l, err := net.Listen("tcp", "0.0.0.0:443")
	if err != nil {
		log.Printf("failed to listen: %s", err)
	}
	l = tls.NewListener(l, tlsCfg)
	if err := http.Serve(l, mux); err != nil {
		log.Printf("couldn't start server: %s", err)
	}
}

func loadClientCertDB(certPath string) ([]*clientCertRecord, error) {
	certPaths := strings.Split(certPath, ",")
	db := make([]*clientCertRecord, 0, len(certPaths))
	for _, cp := range certPaths {
		cp := strings.TrimSpace(cp)
		clientCAPath := filepath.Join(cp, "ca.crt")
		clientCA, err := internal.LoadX509Cert(clientCAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA cert: %s: %w", clientCAPath, err)
		}
		clientCertPath := filepath.Join(cp, "client.crt")
		clientCert, err := internal.LoadX509Cert(clientCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %s: %w", clientCertPath, err)
		}

		clientCertFingerPrint := sha256.Sum256(clientCert.Raw)
		certRecord := clientCertRecord{
			rootCA: clientCA,
			clientCerts: map[string]*x509.Certificate{
				hex.EncodeToString(clientCertFingerPrint[:]): clientCert,
			},
		}
		db = append(db, &certRecord)
	}

	return db, nil
}

func findRootCAFromClientCertDB(clientCert *x509.Certificate) (*x509.Certificate, error) {
	fingerprint := sha256.Sum256(clientCert.Raw)
	for _, r := range clientCertDB {
		if v, ok := r.clientCerts[hex.EncodeToString(fingerprint[:])]; ok {
			return v, nil
		}
	}
	return nil, errors.New("root ca not found")
}
