/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/app"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init in folder",
	Long: `Init in folder. This command should be run in the root of the folder you want to use as a repository. It creates the necessary files to use the repository.
	The user who runs this command is automatically joined in the repository as the owner.`,
	Run: func(cmd *cobra.Command, args []string) {
		// set the private key
		setResult := app.SetPrivateKey(encryptedPrivateKey)
		if !setResult.Ok {
			MarshalOutput(setResult)
			return
		}
		// join the user
		joinRes := app.Join()
		if !joinRes.Ok {
			MarshalOutput(joinRes)
		}
		// init the repo
		res := app.InitRepo()
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	SetRequiredKeyFlag(initCmd)
}
