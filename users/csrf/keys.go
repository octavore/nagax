package csrf

import jose "github.com/square/go-jose"

// load keys from the keystore
// todo: copied from session module, refactor?
func loadKeys(keyFile string, keyStore KeyStore) (interface{}, interface{}, error) {
	privateKeyBytes, _, err := keyStore.LoadPrivateKey(keyFile)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := jose.LoadPrivateKey(privateKeyBytes)
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
	return privateKey, publicKey, nil
}
