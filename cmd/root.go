/*
Copyright © 2024 Mohammad Saadatfar

*/

package cmd

import (
	"ctb-cli/config"
	"ctb-cli/core"
	"ctb-cli/objectstorage/cloud"
	"ctb-cli/repositories"
	"ctb-cli/services/filesyetem_service"
	"ctb-cli/services/key_service"
	"ctb-cli/services/object_service"
	"ctb-cli/services/share_service"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var cfgFile string
var secret string

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
}

func init() {
	cobra.OnInitialize(initConfig)
	addSubCommands()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $USERPROFILE/.ctb/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&secret, "secret", "s", "", "Your secret")
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

var fileSystem *filesyetem_service.FileSystem
var keyStore core.KeyService
var shareService *share_service.Service

func initManagerClient() {
	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := objectstorage.NewDummyClient()

	userId, err := config.Workspace.GetUserId()

	root, _ := config.GetRepoCtbRoot()
	tempRoot, _ := config.GetTempRoot()

	keysPath := CreateAndReturn(filepath.Join(root, "keys"))
	objectPath := CreateAndReturn(filepath.Join(root, "object"))
	recipientsPath := CreateAndReturn(filepath.Join(root, "recipients"))
	filesystemPath := CreateAndReturn(filepath.Join(root, "filesystem"))
	cachePath := CreateAndReturn(filepath.Join(tempRoot, "cache"))

	keyRepository := repositories.NewKeyRepositoryFile(keysPath)
	objectCacheRepository := repositories.NewObjectCacheRepository(cachePath)
	objectRepository := repositories.NewObjectRepository(objectPath)
	recipientRepository := repositories.NewRecipientRepositoryFile(recipientsPath)
	linkRepository := repositories.NewLinkRepository(filesystemPath)

	keyStore = key_service.NewKeyStore(userId, keyRepository)

	path, err := config.Crypto.GetRecoveryPublicCertPath()
	if err != nil {
		return
	}
	err = keyStore.AddRecoveryKey(path)
	if err != nil {
		fmt.Println("Error reading crt:", err)
		return
	}

	objectService := object_service.NewService(keyStore, userId, &objectCacheRepository, &objectRepository, cloudClient)
	shareService = share_service.NewService(recipientRepository, keyStore, linkRepository, &objectService)

	fileSystem = filesyetem_service.NewFileSystem(keyStore, objectService, linkRepository)
}

func CreateAndReturn(path string) string {
	os.MkdirAll(path, os.ModePerm)
	return path
}
