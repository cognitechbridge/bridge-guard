package cmd

import (
	"ctb-cli/app"
	"os"

	"github.com/spf13/cobra"
)

func InitKey() error {
	c := app.SetAndCheckPrivateKey(encpdedPrivateKey)
	if c.Ok {
		return nil // Success, exit function
	} else {
		MarshalOutput(c)
		os.Exit(1) // Failure, exit program
	}
	return nil
}

func SetRequiredKeyFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&encpdedPrivateKey, "key", "k", "", "Your private key")
	c.MarkPersistentFlagRequired("key")
}
