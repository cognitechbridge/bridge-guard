package app

func ChangeSecret(secret string) error {
	err := keyStore.ChangeSecret(secret)
	if err != nil {
		return err
	}
	return nil
}
