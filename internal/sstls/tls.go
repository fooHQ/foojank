package sstls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"time"
)

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func NewCertificateTemplate(org string) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		DNSNames: []string{"localhost"},
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:     true,
	}, nil
}

func EncodeCertificate(cert, parent *x509.Certificate, publicKey crypto.PublicKey, privateKey crypto.PrivateKey) (string, error) {
	b, err := x509.CreateCertificate(rand.Reader, cert, parent, publicKey, privateKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func EncodeKey(privateKey crypto.PrivateKey) (string, error) {
	b, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func DecodeCertificate(s string) ([]byte, *x509.Certificate, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(b)
	if err != nil {
		return nil, nil, err
	}

	return b, cert, nil
}

func DecodeCertificateHandler(s string) func() (*x509.CertPool, error) {
	return func() (*x509.CertPool, error) {
		_, cert, err := DecodeCertificate(s)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		pool.AddCert(cert)
		return pool, nil
	}
}

func DecodeKey(s string) (crypto.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS8PrivateKey(b)
	if err != nil {
		return nil, err
	}

	return key, nil
}
