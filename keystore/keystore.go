package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"
)

func getPrivateKey(fileName string) (*rsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(fileName)
	if err == nil {
		p, _ := pem.Decode(b)
		return x509.ParsePKCS1PrivateKey(p.Bytes)
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	b = x509.MarshalPKCS1PrivateKey(privateKey)
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: b,
	})
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

type KeyStore struct {
	Dir string
}

func (k *KeyStore) LoadPrivateKey(fileName string) ([]byte, *rsa.PrivateKey, error) {
	privateKey, err := getPrivateKey(path.Join(k.Dir, fileName))
	if err != nil {
		return nil, nil, err
	}
	return x509.MarshalPKCS1PrivateKey(privateKey), privateKey, nil
}

func (k *KeyStore) LoadPublicKey(fileName string) ([]byte, error) {
	privateKey, err := getPrivateKey(path.Join(k.Dir, fileName))
	if err != nil {
		return nil, err
	}
	return x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
}
