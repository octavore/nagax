package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
)

const sessionKeyFileName = "session.key"

func getPrivateKey() (*rsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(sessionKeyFileName)
	if err == nil {
		p, _ := pem.Decode(b)
		b, err = x509.DecryptPEMBlock(p, []byte{})
		if err != nil {
			return nil, err
		}
		return x509.ParsePKCS1PrivateKey(b)
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	b = x509.MarshalPKCS1PrivateKey(privateKey)
	block, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", b, []byte{}, x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(sessionKeyFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = pem.Encode(f, block)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

type KeyStore struct{}

func (*KeyStore) LoadPrivateKey() ([]byte, error) {
	privateKey, err := getPrivateKey()
	if err != nil {
		return nil, err
	}
	return x509.MarshalPKCS1PrivateKey(privateKey), nil
}

func (*KeyStore) LoadPublicKey() ([]byte, error) {
	privateKey, err := getPrivateKey()
	if err != nil {
		return nil, err
	}
	return x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
}
