package cmd

import (
	"ctb-cli/app"
	"ctb-cli/prompts"
	"ctb-cli/services/key_service"
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
		err := app.SetAndCheckSecret(secret)
		if err == nil {
			return nil // Success, exit function
		}

		if errors.Is(err, key_service.ErrorInvalidSecret) {
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
