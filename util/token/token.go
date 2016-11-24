package token

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
)

func New64() string {
	b, err := RandN(8)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func RandN(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func New32() string {
	b, err := RandN(5)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(b)
}
