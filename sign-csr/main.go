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
	"os"
	"time"
)

// preparation
//
// $ openssl genrsa -out ca.key 2048
// $ openssl req -x509 -new -nodes -key ca.key -sha256 -days 1024 -out ca.pem -config ca.conf
// $ openssl genrsa -out cert.key 2048
// $ openssl req -new -key cert.key -out cert.csr
//
// run
//
// $ go run main.go -caCert=ca.pem -caPrivateKey=ca.key -csr=cert.csr
//
func main() {
	var (
		caCertPath       string
		caPrivateKeyPath string
		csrPath          string
	)

	flag.StringVar(&caCertPath, "caCert", "", "CA Certificate")
	flag.StringVar(&caPrivateKeyPath, "caPrivateKey", "", "CA Private Key")
	flag.StringVar(&csrPath, "csr", "", "Certificate Signing Request")
	flag.Parse()

	caCertData, err := loadASN1DataFromPEMFile(caCertPath)
	if err != nil {
		panic(err)
	}
	ca, err := x509.ParseCertificate(caCertData)
	if err != nil {
		panic(err)
	}

	caPrivateKeyData, err := loadASN1DataFromPEMFile(caPrivateKeyPath)
	if err != nil {
		panic(err)
	}
	// reference to https://gist.github.com/jshap70/259a87a7146393aab5819873a193b88c
	var caPrivateKey interface{}
	if caPrivateKey, err = x509.ParsePKCS1PrivateKey(caPrivateKeyData); err != nil {
		if caPrivateKey, err = x509.ParsePKCS8PrivateKey(caPrivateKeyData); err != nil {
			panic(err)
		}
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
		SerialNumber:       big.NewInt(r.Int63()),
		Subject:            csr.Subject,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(500 * 24 * time.Hour),
		Extensions:         csr.Extensions,
		ExtraExtensions:    csr.ExtraExtensions,
		Version:            csr.Version,
		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,
		DNSNames:           csr.DNSNames,
		EmailAddresses:     csr.EmailAddresses,
		IPAddresses:        csr.IPAddresses,
		URIs:               csr.URIs,
	}

	certData, err := x509.CreateCertificate(cRand.Reader, &cert, ca, cert.PublicKey, caPrivateKey)
	if err != nil {
		panic(err)
	}

	certBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certData,
	}

	if err := pem.Encode(os.Stdout, &certBlock); err != nil {
		panic(err)
	}
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
