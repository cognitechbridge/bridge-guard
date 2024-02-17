/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/keystore"
	"ctb-cli/prompts"
	"github.com/spf13/cobra"
)

// chsecretCmd represents the chsecret command
var chsecretCmd = &cobra.Command{
	Use:   "chsecret",
	Short: "Change secret",
	Long:  `Change secret`,
	Run: func(cmd *cobra.Command, args []string) {
		secret, err := prompts.SetSecret()
		if err != nil {
			panic(err)
		}
		key, err := keystore.GetKeyFromSecret(secret)
		if err != nil {
			panic(err)
		}
		err = keyStore.ChangeRootKey(key)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(chsecretCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chsecretCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chsecretCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
