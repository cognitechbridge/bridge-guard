/*
Copyright © 2024 Mohammad Saadatfar

*/

package cmd

import (
	"ctb-cli/cmd/file"
	"ctb-cli/config"
	"ctb-cli/db"
	"ctb-cli/encryptor"
	"ctb-cli/file_db/cloud"
	"ctb-cli/filesyetem"
	"ctb-cli/keystore"
	"ctb-cli/manager"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ctb-cli",
	Short: "This is CTB cli tool",
	Long:  `This is CTB cli tool.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func addSubCommands() {
	rootCmd.AddCommand(file.FileCmd)
}

func init() {
	cobra.OnInitialize(initConfig)
	addSubCommands()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $USERPROFILE/.ctb/config.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath("/etc/.ctb/") // path to look for the config file in
		viper.AddConfigPath(".")          // optionally look for config in the working directory
		viper.AddConfigPath(home + "/.ctb")
		viper.SetConfigName("config") // name of config file (without extension)
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	initManagerClient()
}

func initManagerClient() {
	var key encryptor.Key

	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := file_db.NewDummyClient()

	sqlLiteConnection, _ := db.NewSqlLiteConnection()

	keyStore := keystore.NewKeyStore(key, sqlLiteConnection)
	path, err := config.Crypto.GetRecoveryPublicCertPath()
	if err != nil {
		return
	}
	err = keyStore.ReadRecoveryKey(path)
	if err != nil {
		fmt.Println("Error reading crt:", err)
		return
	}

	filesystem := filesyetem.NewPersistFileSystem()

	chunkSize, _ := config.Crypto.GetChunkSize()
	clientId, _ := config.Workspace.GetClientId()
	managerConfig := manager.Config{
		EncryptChunkSize: chunkSize,
		ClientId:         clientId,
	}

	manager.Client.Init(
		managerConfig,
		keyStore,
		filesystem,
		cloudClient,
	)
}
