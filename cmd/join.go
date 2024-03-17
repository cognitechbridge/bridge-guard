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
	Long:  `Join user to the repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		setResult := app.SetPrivateKey(encpdedPrivateKey)
		if !setResult.Ok {
			MarshalOutput(setResult)
			return
		}
		res := app.Join()
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
	SetRequiredKeyFlag(joinCmd)
}
