package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the configuration of the application
type Config struct {
	repoPath   string      // path to the repository
	tempPath   string      // path to the temporary folder of the application
	repoConfig viper.Viper // configuration of the repository
	userConfig viper.Viper // configuration of the user
}

// New returns a new Config
func New(repoPath string, tempPath string, cfgFile string) (*Config, error) {
	repoConfig := viper.New()
	repoConfig.SetConfigName("ctb")
	repoConfig.SetConfigType("yaml")
	repoConfig.AddConfigPath(repoPath)

	userConfig := viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		userConfig.SetConfigFile(cfgFile)
	} else {
		// Search config in different directories
		viper.AddConfigPath("/etc/.ctb/") // path to look for the config file in
		viper.SetConfigName("config")     // name of config file (without extension)
		viper.SetConfigType("yaml")

		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			viper.AddConfigPath(home + "/.ctb")
		}
	}

	// Read in the config file and ignore errors! User can check if the file exists using IsRepositoryConfigExists method
	_ = userConfig.ReadInConfig()
	_ = repoConfig.ReadInConfig()

	return &Config{
		repoPath:   repoPath,
		tempPath:   tempPath,
		repoConfig: *repoConfig,
		userConfig: *userConfig,
	}, nil
}

// GetRepoRoot returns the root path of the repository.
func (c *Config) GetRepoRoot() (string, error) {
	return c.repoPath, nil
}

// GetTempRoot returns the root path of the temporary folder.
func (c *Config) GetTempRoot() (string, error) {
	return c.tempPath, nil
}

// GetRepoCtbRoot returns the root path of the repository.
func (c *Config) GetRepoCtbRoot() (string, error) {
	return c.repoPath, nil
}

// InitRepoConfig generates the configuration file for the repository.
func (c *Config) InitRepoConfig(repoId string) error {
	// Set the default values for the configuration
	c.repoConfig.SetConfigFile(filepath.Join(c.repoPath, "ctb.yaml"))
	c.repoConfig.Set("id", repoId)
	err := c.repoConfig.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

// IsRepositoryConfigExists checks if the repository configuration exists.
func (c *Config) IsRepositoryConfigExists() bool {
	err := c.repoConfig.ReadInConfig()
	return err == nil
}

// GetRepoId returns the id of the repository.
func (c *Config) GetRepoId() string {
	return c.repoConfig.GetString("id")
}
