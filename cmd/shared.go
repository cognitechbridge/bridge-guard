package cmd

import (
	"ctb-cli/keystore"
	"ctb-cli/prompts"
	"errors"
	"github.com/fatih/color"
)

func InitSecret() error {
	needSecret := secret == "" // Determine if we need to prompt for the secret

	for {
		if needSecret {
			var err error
			secret, err = prompts.GetSecret()
			if err != nil {
				return err // If there's an error getting the secret, return immediately
			}
		}
		keyStore.SetSecret(secret)
		err := keyStore.LoadKeys()
		if err == nil {
			return nil // Success, exit function
		}

		if errors.Is(err, keystore.ErrorInvalidSecret) {
			// Notify user of invalid secret
			c := color.New(color.FgRed, color.Bold)
			_, _ = c.Println("Invalid secret. Try again")

			if !needSecret {
				return err
			}
		} else {
			return err // For any other error, return immediately
		}
	}
}
