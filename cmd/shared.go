package cmd

import (
	"github.com/spf13/cobra"
)

// SetRequiredKeyFlag sets the required 'key' flag for a command.
func SetRequiredKeyFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&encryptedPrivateKey, "key", "k", "", "Your private key. Required.")
	err := c.MarkPersistentFlagRequired("key")
	if err != nil {
		panic(err)
	}
}
