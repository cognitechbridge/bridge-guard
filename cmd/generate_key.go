/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// genrateKeyCmd represents the generateKey command
var generateKeyCmd = &cobra.Command{
	Use:   "generate-key",
	Short: "Generate a new private key for user",
	Long: `Generate a new private key for user and return it as a string. The key is used to encrypt and decrypt data keys and vaults.
	This funtion doesnt affect the state of the repository and only returns the key. The key is not stored in the repository.
	You can use join command to join the repository and store the corresponding public key in the repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		res := ctbApp.GenerateUserKey()
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(generateKeyCmd)
}
