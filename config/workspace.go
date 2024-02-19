package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type workspace struct{}

var Workspace = workspace{}

func (*workspace) GetClientId() (string, error) {
	path := "workspace.client-id"
	res := viper.GetString(path)
	if res == "" {
		return "", fmt.Errorf("workspace.client-id not found")
	}
	return res, nil
}

func (*workspace) SetClientId(clientId string) error {
	viper.Set("workspace.client-id", clientId)
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

func (*workspace) GetEmail() (string, error) {
	path := "workspace.email"
	res := viper.GetString(path)
	if res == "" {
		return "", fmt.Errorf("workspace.email not found")
	}
	return res, nil
}

func (*workspace) SetEmail(clientId string) error {
	viper.Set("workspace.email", clientId)
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
