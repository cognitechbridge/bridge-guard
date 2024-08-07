/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// pubkeyCmd represents the pubkey command
var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "Generate a public key from a private key",
	Long:  `Generate a public key from a private key.`,
	Run: func(cmd *cobra.Command, args []string) {
		res := ctbApp.GetPubkey(encryptedPrivateKey)
		MarshalOutput(res)
	},
}

func init() {
	RootCmd.AddCommand(pubkeyCmd)
	SetRequiredKeyFlag(pubkeyCmd)
}
