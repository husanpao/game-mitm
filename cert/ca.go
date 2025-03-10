package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// CA represents a Certificate Authority
type CA struct {
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
	CertPath    string
	KeyPath     string
}

// LoadOrCreateCA loads an existing CA certificate or creates a new one if it doesn't exist
func LoadOrCreateCA(caDir string) (*CA, error) {
	certPath := filepath.Join(caDir, "ca.crt")
	keyPath := filepath.Join(caDir, "ca.key")

	// Check if the CA certificate and key already exist
	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			return loadCA(certPath, keyPath)
		}
	}

	// Create a new CA
	return createCA(certPath, keyPath)
}

// loadCA loads the CA certificate and private key from files
func loadCA(certPath, keyPath string) (*CA, error) {

	// Read CA certificate
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to parse CA certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %v", err)
	}

	// Read CA private key
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA private key: %v", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to parse CA private key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA private key: %v", err)
	}

	return &CA{
		Certificate: cert,
		PrivateKey:  key,
		CertPath:    certPath,
		KeyPath:     keyPath,
	}, nil
}

// createCA creates a new CA certificate and private key
func createCA(certPath, keyPath string) (*CA, error) {
	log.Println("Creating new CA certificate")

	// Generate private key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Generate certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:    "MITM Proxy CA",
			Organization:  []string{"MITM Proxy"},
			Country:       []string{"CN"},
			Province:      []string{"Anywhere"},
			Locality:      []string{"Anywhere"},
			StreetAddress: []string{"Anywhere"},
			PostalCode:    []string{"000000"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create CA certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %v", err)
	}

	// Parse CA certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %v", err)
	}

	// Save CA certificate to file
	certOut, err := os.Create(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate file: %v", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certOut.Close()

	// Save CA private key to file
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA private key file: %v", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	keyOut.Close()

	return &CA{
		Certificate: cert,
		PrivateKey:  key,
		CertPath:    certPath,
		KeyPath:     keyPath,
	}, nil
}
