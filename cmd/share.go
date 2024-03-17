/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/app"

	"github.com/spf13/cobra"
)

// shareCmd represents the share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share files with other users",
	Long:  `Share files with other users`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		InitKey()
		pattern := args[0]
		recipient, _ := cmd.Flags().GetString("recipient")
		res := app.Share(pattern, recipient)
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(shareCmd)
	SetRequiredKeyFlag(shareCmd)
	shareCmd.PersistentFlags().StringP("recipient", "r", "", "recipient email address")
	shareCmd.MarkPersistentFlagRequired("recipient")
}
