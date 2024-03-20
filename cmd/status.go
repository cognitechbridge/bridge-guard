/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/app"

	"github.com/spf13/cobra"
)

// stausCmd represents the staus command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of the repository.",
	Long: `Get the status of the repository. It checks if the repository is valid and if the user has joined.
	Returns an AppResult with the repository status.
	You can use the 'key' flag to pass your private key. If you don't pass it, the joined status will be false.`,
	Run: func(cmd *cobra.Command, args []string) {
		res := app.GetStatus(encryptedPrivateKey)
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.PersistentFlags().StringVarP(&encryptedPrivateKey, "key", "k", "", "Your private key. Optional.")
}
