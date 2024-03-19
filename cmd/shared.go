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
func initKey() {
	c := app.SetAndCheckPrivateKey(encryptedPrivateKey)
	if c.Ok {
		return // Success, exit function
	} else {
		MarshalOutput(c)
		os.Exit(1) // Failure, exit program
	}
}

// SetRequiredKeyFlag sets the required 'key' flag for a command.
func SetRequiredKeyFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&encryptedPrivateKey, "key", "k", "", "Your private key. Required.")
	err := c.MarkPersistentFlagRequired("key")
	if err != nil {
		panic(err)
	}
}
