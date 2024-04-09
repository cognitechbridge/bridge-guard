/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// listAccessCmd represents the listAccess command
var listAccessCmd = &cobra.Command{
	Use:   "list-access",
	Short: "List access to a file or directory",
	Long: `This command lists the access to a file or directory located at the specified path.
	The access list includes the public keys of users who have access to the file or directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		res := ctbApp.ListAccess(path)
		MarshalOutput(res)
	},
}

func init() {
	rootCmd.AddCommand(listAccessCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listAccessCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listAccessCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
