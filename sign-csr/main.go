package main

import (
	cRand "crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"time"
)

func main() {
	var (
		caCertPath string
		caPrivateKeyPath string
		csrPath string
	)

	flag.StringVar(&caCertPath, "caCert", "", "CA Certificate")
	flag.StringVar(&caPrivateKeyPath, "caPrivateKey", "", "CA Private Key")
	flag.StringVar(&csrPath, "car", "", "Certificate Signing Request")
	flag.Parse()

	caCertData, err := loadASN1DataFromPEMFile(caCertPath)
	if err != nil {
		panic(err)
	}
	ca, err := x509.ParseCertificate(caCertData)
	if err != nil {
		panic(err)
	}

	caPrivateKeyData, err := loadASN1DataFromPEMFile(caCertPath)
	if err != nil {
		panic(err)
	}

	csrData, err := loadASN1DataFromPEMFile(csrPath)
	if err != nil {
		panic(err)
	}

	csr, err := x509.ParseCertificateRequest(csrData)
	if err != nil {
		panic(err)
	}

	if err := csr.CheckSignature(); err != nil {
		panic(err)
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	cert := x509.Certificate{
		SerialNumber:                big.NewInt(r.Int63()),
		Subject:                     csr.Subject,
		NotBefore:                   time.Now(),
		NotAfter:                    time.Now().Add(500*24*time.Hour),
		Extensions:                  csr.Extensions,
		ExtraExtensions:             csr.ExtraExtensions,
	}

	if err := x509.CreateCertificate(cRand.Reader, )
}

func loadASN1DataFromPEMFile(pemFilePath string) ([]byte, error) {
	pemData, err := ioutil.ReadFile(pemFilePath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read PEM file: %s: %w", pemFilePath, err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("invalid pem file: %s", pemFilePath)
	}
	return block.Bytes, nil
}
