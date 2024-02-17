package prompts

import (
	"github.com/erikgeiser/promptkit/textinput"
)

func GetSecret() (string, error) {
	input := textinput.New("Enter your secret (passphrase):")
	input.Placeholder = "..........."
	input.Validate = validate
	input.Hidden = true
	input.Template += validatedTextboxTemplate
	pass, err := input.RunPrompt()
	if err != nil {
		return "", err
	}
	return pass, nil
}

func validate(s string) error {
	return nil
}
