package cmd

import (
	"ctb-cli/app"
	"ctb-cli/core"
	"ctb-cli/prompts"
	"ctb-cli/services/key_service"
	"encoding/json"
	"errors"
	"fmt"
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
		res := app.SetAndCheckSecret(secret)
		if res.Ok {
			return nil // Success, exit function
		}

		if errors.Is(res.Err, key_service.ErrorInvalidSecret) {
			// Notify user of invalid secret
			c := color.New(color.FgRed, color.Bold)
			_, _ = c.Println("Invalid secret. Try again")

			if !needSecret {
				return res.Err
			}
		} else {
			return res.Err // For any other error, return immediately
		}
	}
}

func MarshalOutput(result core.AppResult) {
	res, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(res))
}
