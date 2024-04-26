package config_service

import (
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the configuration of the application
type ConfigService struct {
	rootPath string
}

// New returns a new Config
func New(rootPath string) *ConfigService {
	return &ConfigService{
		rootPath: rootPath,
	}
}

// InitConfig generates the configuration file for the repository.
func (c *ConfigService) InitConfig(path string) error {
	configPath := c.getConfigPath(path)
	cfg := viper.New()
	// Set the default values for the configuration
	cfg.SetConfigFile(filepath.Join(configPath, "config.yaml"))
	cfg.Set("version", 1)
	err := cfg.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

// IsRepositoryConfigExists checks if the repository configuration exists.
func (c *ConfigService) IsRepositoryConfigExists(path string) bool {
	err := c.getConfig(path).ReadInConfig()
	return err == nil
}

// GetRepoVersion returns the version of the repository.
func (c *ConfigService) GetRepoVersion(path string) string {
	return c.getConfig(path).GetString("version")
}

// GetRepoConfig returns the configuration of the path.
func (c *ConfigService) getConfig(path string) *viper.Viper {
	configPath := c.getConfigPath(path)
	cfg := viper.New()
	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(configPath)

	_ = cfg.ReadInConfig()

	return cfg
}

func (c *ConfigService) getConfigPath(path string) string {
	return filepath.Join(c.rootPath, path, ".meta")
}
