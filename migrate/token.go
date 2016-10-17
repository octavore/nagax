package migrate

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

func randomToken() string {
	b := make([]byte, 5)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return strings.ToLower(base32.StdEncoding.EncodeToString(b))
}
