package core

func GenerateKey() (*KeyInfo, error) {
	key := NewKeyFromRand()
	keyId, err := NewUid()
	if err != nil {
		return nil, err
	}
	keyInfo := NewKeyInfo(keyId, key[:])
	return &keyInfo, nil
}
