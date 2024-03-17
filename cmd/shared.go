package cmd

import (
	"ctb-cli/app"
	"os"

	"github.com/spf13/cobra"
)

// initKey initializes the private key.
// It sets and checks the private key that is passed as a 'key' flag.
// If the operation is successful, it returns nil.
// If the operation fails, it marshals the output and exits the program with a status code of 1.
func initKey() error {
	c := app.SetAndCheckPrivateKey(encryptedPrivateKey)
	if c.Ok {
		return nil // Success, exit function
	} else {
		MarshalOutput(c)
		os.Exit(1) // Failure, exit program
	}
	return nil
}

// SetRequiredKeyFlag sets the required 'key' flag for a command.
func SetRequiredKeyFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&encryptedPrivateKey, "key", "k", "", "Your private key. Required.")
	c.MarkPersistentFlagRequired("key")
}
