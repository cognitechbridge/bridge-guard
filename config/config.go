package config

// Config represents the configuration of the application
type Config struct {
	repoPath string // path to the repository
	tempPath string // path to the temporary folder of the application
}

// New returns a new Config
func New(repoPath string, tempPath string, cfgFile string) (*Config, error) {
	return &Config{
		repoPath: repoPath,
		tempPath: tempPath,
	}, nil
}

// GetTempRoot returns the root path of the temporary folder.
func (c *Config) GetTempRoot() (string, error) {
	return c.tempPath, nil
}

// GetRepoCtbRoot returns the root path of the repository.
func (c *Config) GetRepoCtbRoot() (string, error) {
	return c.repoPath, nil
}
