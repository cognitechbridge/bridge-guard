/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"ctb-cli/manager"
	"fmt"
	"github.com/spf13/cobra"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads a file to cloud",
	Long:  `Uploads a file to cloud`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		name, _ := cmd.Flags().GetString("name")

		uploader := manager.Client.NewUploader(
			file,
			name,
		)
		res, err := uploader.Upload()
		if err != nil {
			fmt.Printf("Error uploading: %v", err)
			return
		}
		fmt.Printf("Upload completed: %s\n", res)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringP("name", "n", "", "name on cloud")
	uploadCmd.Flags().StringP("file", "f", "", "File to upload")
	_ = uploadCmd.MarkFlagRequired("path")
	_ = uploadCmd.MarkFlagRequired("file")
}
