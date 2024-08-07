/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// unshareCmd represents the unshare command
var unshareCmd = &cobra.Command{
	Use:   "unshare",
	Short: "Unshare files with other users",
	Long:  `This command unshares file or directory with the specified path with the given public key.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		recipient, _ := cmd.Flags().GetString("recipient")
		res := ctbApp.Unshare(path, recipient)
		MarshalOutput(res)
	},
}

func init() {
	RootCmd.AddCommand(unshareCmd)
	unshareCmd.PersistentFlags().StringP("recipient", "r", "", "recipient public key. Required.")
	err := unshareCmd.MarkPersistentFlagRequired("recipient")
	if err != nil {
		panic(err)
	}
}
