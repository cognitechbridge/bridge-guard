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
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")
		force, _ := cmd.Flags().GetBool("force")

		uploader := manager.Client.NewUploader(path, name, force)
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
	uploadCmd.Flags().StringP("path", "p", "", "path to file to upload")
	uploadCmd.Flags().BoolP("force", "f", false, "force")
	_ = uploadCmd.MarkFlagRequired("path")
	_ = uploadCmd.MarkFlagRequired("file")
}
