package internal

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// LoadX509Cert load x509.Certificate. The file at certPath should be PEM encoded DER.
func LoadX509Cert(certPath string) (*x509.Certificate, error) {
	pemData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read PEM file: %s: %w", certPath, err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("invalid pem file: %s", certPath)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse as certificate: %w", err)
	}
	return cert, nil
}

func LoadTLSCert(certDirPath, certFileName, keyFileName string) (tls.Certificate, error) {
	certFilePath := filepath.Join(certDirPath, certFileName)
	keyFilePath := filepath.Join(certDirPath, keyFileName)
	return tls.LoadX509KeyPair(certFilePath, keyFilePath)
}
