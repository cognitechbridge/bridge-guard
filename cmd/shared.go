package cmd

import (
	"ctb-cli/app"
)

func InitKey() error {
	c := app.SetAndCheckPrivateKey(encpdedPrivateKey)
	if c.Ok {
		return nil // Success, exit function
	} else {
		MarshalOutput(c)
		panic(nil) // Exit program with panic
	}
}
