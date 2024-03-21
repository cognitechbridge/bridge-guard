package config

import "github.com/spf13/viper"

type Config struct {
	repoPath   string
	tempPath   string
	repoConfig viper.Viper
}

// New returns a new Config
func New(repoPath string, tempPath string) (*Config, error) {
	crepoConfig := viper.New()
	crepoConfig.SetConfigName("ctb")
	crepoConfig.SetConfigType("yaml")
	crepoConfig.AddConfigPath(repoPath)
	return &Config{
		repoPath:   repoPath,
		tempPath:   tempPath,
		repoConfig: *crepoConfig,
	}, nil
}

func (c *Config) GetRepoRoot() (string, error) {
	return c.repoPath, nil
}

func (c *Config) GetTempRoot() (string, error) {
	return c.tempPath, nil
}

func (c *Config) GetRepoCtbRoot() (string, error) {
	return c.repoPath, nil
}
