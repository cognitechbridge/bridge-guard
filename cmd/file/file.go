/*
Copyright © 2024 Mohammad Saadatfar

*/

package file

import (
	"github.com/spf13/cobra"
)

// FileCmd represents the file command
var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "File commands root",
	Long:  `File commands root`,
}

func init() {

	FileCmd.PersistentFlags().BoolP("force", "f", false, "force")
	FileCmd.PersistentFlags().BoolP("recursive", "r", false, "recursive")
}
