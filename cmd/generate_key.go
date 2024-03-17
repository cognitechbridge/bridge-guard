/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/app"

	"github.com/spf13/cobra"
)

// genrateKeyCmd represents the genrateKey command
var generateKeyCmd = &cobra.Command{
	Use:   "generate-key",
	Short: "Generate a new private key for user",
	Long:  `Generate a new private key for user.`,
	Run: func(cmd *cobra.Command, args []string) {
		res := app.GenerateUserKey()
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(generateKeyCmd)
}
