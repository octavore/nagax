package token

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

var encoder = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz")

func New64() string {
	return New(8)
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
	return New(5)
}

func New(n int) string {
	b, err := RandN(n)
	if err != nil {
		panic(err)
	}
	return strings.TrimRight(encoder.EncodeToString(b), "=")
}
