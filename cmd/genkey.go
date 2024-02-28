/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/rand"
	"ctb-cli/config"
	"ctb-cli/prompts"
	"ctb-cli/core"
	"fmt"
	"github.com/spf13/cobra"
	"io"
)

// genkeyCmd represents the genkey command
var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate user key pairs",
	Long:  `Generate user key pairs`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		//Email
		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			email, err = prompts.GetEmail()
			if err != nil {
				panic(err)
			}
		}

		//Secret
		secret, err := prompts.NewSecret()
		if err != nil {
			panic(err)
		}

		//Save email to config
		err = config.Workspace.SetEmail(email)
		if err != nil {
			panic(err)
		}

		//Generate random use id
		bytes := make([]byte, 128/8)
		_, err = io.ReadFull(rand.Reader, bytes)
		if err != nil {
			panic(err)
		}
		userId, err := core.EncodeUid(bytes)
		if err != nil {
			panic(err)
		}

		//Save user id to config and keystore
		err = config.Workspace.SetUserId(userId)
		keyStore.SetUserId(userId)
		if err != nil {
			panic(err)
		}

		keyStore.SetSecret(secret)
		if err != nil {
			panic(err)
		}
		err = keyStore.GenerateUserKeys()
		if err != nil {
			panic(err)
		}

		publicKey, err := keyStore.GetPublicKey()
		if err != nil {
			panic(err)
		}
		recipient, err := core.NewRecipient(email, publicKey, userId)
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
	genkeyCmd.PersistentFlags().StringP("email", "e", "", "email")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genkeyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
