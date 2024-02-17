package prompts

import (
	"errors"
	"github.com/erikgeiser/promptkit/textinput"
	"regexp"
)

var (
	validatedTextboxTemplate = `
	{{- if .ValidationError -}}
		{{- print " " (Foreground "1" .ValidationError.Error) -}}
	{{- end -}}`
)

func SetSecret() (string, error) {

	input := textinput.New("Choose a new secret (passphrase):")
	input.Placeholder = "make it strong!"
	input.Validate = validatePassword
	input.Hidden = true
	input.Template += validatedTextboxTemplate
	pass, err := input.RunPrompt()
	if err != nil {
		return "", err
	}

	input = textinput.New("Repeat your new secret:")
	input.Validate = validateRepeat(pass)
	input.Hidden = true
	input.Template += validatedTextboxTemplate
	_, err = input.RunPrompt()
	if err != nil {
		return "", err
	}
	return pass, nil
}

func validatePassword(s string) error {
	if len(s) < 6 {
		return errors.New("the secret must be at least 6 characters long")
	}

	// Define regular expressions for each type of character
	upperCase := regexp.MustCompile(`[A-Z]`)
	lowerCase := regexp.MustCompile(`[a-z]`)
	number := regexp.MustCompile(`[0-9]`)
	specialChar := regexp.MustCompile(`[\W_]`) // Matches any non-word character plus underscore

	// Count how many of the character types are present
	count := 0
	if upperCase.FindStringIndex(s) != nil {
		count++
	}
	if lowerCase.FindStringIndex(s) != nil {
		count++
	}
	if number.FindStringIndex(s) != nil {
		count++
	}
	if specialChar.FindStringIndex(s) != nil {
		count++
	}

	// Check if at least 3 out of 4 character types are present
	if count < 3 {
		return errors.New("the secret must include at least 3 out of 4 character types: uppercase, lowercase, number, and special characters")
	}

	return nil
}

func validateRepeat(first string) func(string) error {
	return func(sec string) error {
		if sec != first {
			return errors.New("passphrases do not match")
		}
		return nil
	}
}
