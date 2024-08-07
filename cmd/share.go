/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/services/key_service"

	"github.com/spf13/cobra"
)

// shareCmd represents the share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share files with other users",
	Long: `This command shares file or directory with the specified path with the given public key.
	The files are shared with the user who has the corresponding private key.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		recipient, _ := cmd.Flags().GetString("recipient")
		if join, _ := cmd.Flags().GetBool("join"); join {
			joinRes := ctbApp.JoinByUserId(recipient)
			if !joinRes.Ok && joinRes.Err != key_service.ErrUserAlreadyJoined {
				MarshalOutput(joinRes)
				return
			}
		}
		res := ctbApp.Share(path, recipient, encryptedPrivateKey)
		MarshalOutput(res)
	},
}

func init() {
	RootCmd.AddCommand(shareCmd)
	SetRequiredKeyFlag(shareCmd)
	shareCmd.PersistentFlags().StringP("recipient", "r", "", "recipient public key. Required.")
	shareCmd.Flags().BoolP("join", "j", false, "Join the user if not already joined.")
	err := shareCmd.MarkPersistentFlagRequired("recipient")
	if err != nil {
		panic(err)
	}
}
