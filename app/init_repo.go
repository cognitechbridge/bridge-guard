package app

func InitRepo() error {
	return fileSystem.CreateVaultInPath("/")
}
