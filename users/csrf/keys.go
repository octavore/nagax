package csrf

import jose "github.com/square/go-jose"

// load keys from the keystore
// todo: copied from session module, refactor?
func loadKeys(keyFile string, keyStore KeyStore) (jose.Encrypter, interface{}, error) {
	privateKey, _, err := keyStore.LoadPrivateKey(keyFile)
	if err != nil {
		return nil, nil, err
	}
	decryptionKey, err := jose.LoadPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	pub, err := keyStore.LoadPublicKey(keyFile)
	if err != nil {
		return nil, nil, err
	}
	publicKey, err := jose.LoadPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}

	encrypter, err := jose.NewEncrypter(keyAlgorithm, contentEncryption, publicKey)
	if err != nil {
		return nil, nil, err
	}

	return encrypter, decryptionKey, err
}
