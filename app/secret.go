package app

func SetAndCheckSecret(secret string) error {
	keyStore.SetSecret(secret)
	err := keyStore.LoadKeys()
	return err
}
