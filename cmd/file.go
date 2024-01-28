/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// FileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "File commands root",
	Long:  `File commands root`,
}

func init() {
	rootCmd.AddCommand(fileCmd)

	fileCmd.PersistentFlags().BoolP("force", "f", false, "force")
	fileCmd.PersistentFlags().BoolP("recursive", "r", false, "recursive")
}
