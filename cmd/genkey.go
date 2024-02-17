/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ctb-cli/prompts"
	"fmt"
	"github.com/spf13/cobra"
)

// genkeyCmd represents the genkey command
var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate client key pairs",
	Long:  `Generate client key pairs`,
	Run: func(cmd *cobra.Command, args []string) {
		secret, err := prompts.SetSecret()
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
