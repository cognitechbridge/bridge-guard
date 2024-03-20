/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/app"

	"github.com/spf13/cobra"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join user to the repository",
	Long: `Join user to the repository. This command join the current user to the repository by storing the corresponding public key in the repository. 
	Use generate-key command to generate the private key.`,
	Run: func(cmd *cobra.Command, args []string) {
		// join the user
		res := app.Join(encryptedPrivateKey)
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
	SetRequiredKeyFlag(joinCmd)
}
