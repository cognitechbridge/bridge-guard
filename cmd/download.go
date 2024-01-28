/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/manager"
	"fmt"

	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a file from cloud",
	Long:  `Download a file from cloud`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")

		downloader := manager.Client.NewDownloader(
			path,
			name,
		)
		err := downloader.Download()
		if err != nil {
			fmt.Printf("Error downloading:%v", err)
			return
		}
		fmt.Printf("Download completed. \n")
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringP("name", "n", "", "name on cloud")
	downloadCmd.Flags().StringP("path", "p", "", "path to download location")
	_ = downloadCmd.MarkFlagRequired("name")
	_ = downloadCmd.MarkFlagRequired("path")
}
