package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

// go run main.go cert
func main() {
	var (
		certPath, rootCAPath, dummyCAPath string
	)

	flag.StringVar(&certPath, "cert", "", "certificate")
	flag.StringVar(&rootCAPath, "rootCA", "", "root CA")
	flag.StringVar(&dummyCAPath, "dummyCA", "", "dummy root CA")
	flag.Parse()

	cert, err := loadCertFromPEMFile(certPath)
	if err != nil {
		panic(err)
	}
	log.Printf("cert: %+v", cert.Subject)

	rootCA, err := loadCertFromPEMFile(rootCAPath)
	if err != nil {
		panic(err)
	}
	log.Printf("rootCA: %+v", rootCA.Subject)
	rootCAPool := x509.NewCertPool()
	rootCAPool.AddCert(rootCA)

	verifyOpts := x509.VerifyOptions{
		Roots:         rootCAPool,
		Intermediates: x509.NewCertPool(),
	}

	chain, err := cert.Verify(verifyOpts)
	if err != nil {
		panic(fmt.Errorf("unexpected verify error: %w", err))
	}
	log.Printf("succeeded to veryify cert")
	if len(chain) != 1 {
		panic(fmt.Errorf("unexpected number of cert chain: %d", len(chain)))
	}
	if len(chain[0]) != 2 {
		panic(fmt.Errorf("unexpected number of cert chain: %d", len(chain[0])))
	}
	if !cert.Equal(chain[0][0]) {
		panic(fmt.Errorf("unexpected cert: %s", chain[0][1].Subject))
	}
	if !rootCA.Equal(chain[0][1]) {
		panic(fmt.Errorf("unexpected cert: %s", chain[0][1].Subject))
	}

	dummyCA, err := loadCertFromPEMFile(dummyCAPath)
	if err != nil {
		panic(err)
	}
	log.Printf("dummyCA: %+v", dummyCA.Subject)
	dummyCAPool := x509.NewCertPool()
	dummyCAPool.AddCert(dummyCA)

	verifyOpts = x509.VerifyOptions{
		Roots:         dummyCAPool,
		Intermediates: x509.NewCertPool(),
	}

	chain, err = cert.Verify(verifyOpts)
	if err == nil {
		panic(fmt.Errorf("expected error didn't happen"))
	}
	if chain != nil {
		panic(fmt.Errorf("unexpected cert chain was returned: %+v", chain))
	}
	log.Printf("expected error happened: %s", err)
}

func loadCertFromPEMFile(pemFilePath string) (*x509.Certificate, error) {
	pemData, err := ioutil.ReadFile(pemFilePath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read PEM file: %s: %w", pemFilePath, err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("invalid pem file: %s", pemFilePath)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse as certificate: %w", err)
	}
	return cert, nil
}
