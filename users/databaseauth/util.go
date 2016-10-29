package databaseauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

func HashPassword(password, salt string) string {
	h := pbkdf2.Key([]byte(password), []byte(salt), 8192, sha256.Size, sha256.New)
	return fmt.Sprintf("%s#%s", salt, base64.StdEncoding.EncodeToString(h))
}

func AuthenticatePassword(given, actual string) bool {
	salt := strings.SplitN(given, "#", 2)[0]
	hash2 := []byte(HashPassword(actual, salt))
	return subtle.ConstantTimeCompare([]byte(given), hash2) == 1
}
