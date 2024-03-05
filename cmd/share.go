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
		err := InitSecret()
		if err != nil {
			panic(err)
		}
		pattern := args[0]
		recipient, _ := cmd.Flags().GetString("recipient")
		res := app.Share(pattern, recipient)
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(shareCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// shareCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	shareCmd.PersistentFlags().StringP("recipient", "r", "", "recipient email address")
	shareCmd.MarkPersistentFlagRequired("recipient")
}
