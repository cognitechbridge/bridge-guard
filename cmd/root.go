/*
Copyright © 2024 Mohammad Saadatfar

*/

package cmd

import (
	"ctb-cli/app"
	"ctb-cli/config"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

var cfgFile string
var repoPath string
var encryptedPrivateKey string
var output outputEnum = outputEnumText

var ctbApp app.App

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

	rootCmd.PersistentFlags().StringVarP(&repoPath, "path", "p", "", "path to the repository")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $USERPROFILE/.ctb/config.yaml)")
	rootCmd.PersistentFlags().VarP(&output, "output", "o", `Output format. allowed: "json", "text", "yaml", and "xml"`)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Get the repository root path
	var repoRootPath string
	if repoPath != "" { // if the path is provided using the flag
		repoRootPath = repoPath
	} else { // if the path is not provided using the flag
		var err error
		repoRootPath, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}
	//Prepare Logger
	prepareLogger(getLogPath())
	// Create the config
	cfg, err := config.New(
		repoRootPath,
		getTempPath(),
		cfgFile,
	)
	if err != nil {
		panic(err)
	}
	// Create the app
	ctbApp = app.New(*cfg)
}

func prepareLogger(logpath string) {
	log.SetOutput(&lumberjack.Logger{
		Filename:   logpath, // File to log to
		MaxSize:    10,      // Size in MB before the file is rotated
		MaxBackups: 30,      // Max number of old log files to keep
		MaxAge:     7,       // Max age in days to keep a log file
		Compress:   true,    // Compress old log files
	})
	log.Info("App Started")
}

func getTempPath() string {
	tempPath := filepath.Join(os.TempDir(), ".ctb")
	return tempPath
}

func getLogPath() string {
	if runtime.GOOS == "windows" {
		homeDir := os.Getenv("UserProfile")
		if homeDir == "" {
			panic("HOME environment variable not set")
		}
		logDir := filepath.Join(homeDir, ".cognitechbridge", "logs")
		logPath := filepath.Join(logDir, "client.log")
		return logPath
	} else if runtime.GOOS == "linux" {
		logDir := filepath.Join("/var/log", "cognitechbridge")
		logPath := filepath.Join(logDir, "client.log")
		return logPath
	} else {
		panic("Unsupported OS")
	}
}
