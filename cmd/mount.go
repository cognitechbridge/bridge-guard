/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/fuse"
	"ctb-cli/manager"
	"github.com/spf13/cobra"
)

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount",
	Long:  `mount`,
	Run: func(cmd *cobra.Command, args []string) {
		ch := make(chan string, 3)
		go manager.Client.Filesystem.UploadQueue.ProcessRoutine(ch)
		go manager.Client.Uploader.UploadRoutine(ch)
		ctbFuse := fuse.NewMemfs(manager.Client.Filesystem)
		ctbFuse.Mount()
	},
}

func init() {
	rootCmd.AddCommand(mountCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
