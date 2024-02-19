/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/rand"
	"ctb-cli/config"
	"ctb-cli/crypto/bech32"
	"ctb-cli/prompts"
	"ctb-cli/types"
	"fmt"
	"github.com/spf13/cobra"
	"io"
)

// genkeyCmd represents the genkey command
var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate client key pairs",
	Long:  `Generate client key pairs`,
	Run: func(cmd *cobra.Command, args []string) {
		//Prompts
		email, err := prompts.GetEmail()
		if err != nil {
			panic(err)
		}
		secret, err := prompts.NewSecret()
		if err != nil {
			panic(err)
		}

		//Save email to config
		err = config.Workspace.SetEmail(email)
		if err != nil {
			panic(err)
		}

		//Generate random client id
		bytes := make([]byte, 128/8)
		_, err = io.ReadFull(rand.Reader, bytes)
		if err != nil {
			panic(err)
		}
		clientId, err := bech32.Encode("ctb-add", bytes)
		if err != nil {
			panic(err)
		}

		//Save client id to config
		err = config.Workspace.SetClientId(clientId)
		if err != nil {
			panic(err)
		}

		err = keyStore.SetSecret(secret)
		if err != nil {
			panic(err)
		}
		err = keyStore.GenerateClientKeys()
		if err != nil {
			panic(err)
		}

		publicKey, err := keyStore.GetPublicKey()
		if err != nil {
			panic(err)
		}
		recipient, err := types.NewRecipient(email, publicKey, clientId)
		if err != nil {
			panic(err)
		}
		err = shareService.SaveRecipient(recipient)
		if err != nil {
			panic(err)
		}

		fmt.Println("Code generated successfully")
	},
}

func init() {
	rootCmd.AddCommand(genkeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genkeyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genkeyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
