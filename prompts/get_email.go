package prompts

import (
	"errors"
	"github.com/erikgeiser/promptkit/textinput"
	"regexp"
)

func GetEmail() (string, error) {

	input := textinput.New("Enter your email:")
	input.Placeholder = "aa@email.com"
	input.Validate = isValidEmail
	input.Template += validatedTextboxTemplate
	pass, err := input.RunPrompt()
	if err != nil {
		return "", err
	}

	return pass, nil
}

func isValidEmail(email string) error {
	// Regular expression for validating an email
	var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("email is not valid")
	}
	return nil
}
