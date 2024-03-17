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
	Long:  `Init in folder`,
	Run: func(cmd *cobra.Command, args []string) {
		initKey()
		res := app.InitRepo()
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	SetRequiredKeyFlag(initCmd)
}
