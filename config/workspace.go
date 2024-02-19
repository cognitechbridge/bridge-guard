package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type workspace struct{}

var Workspace = workspace{}

func (*workspace) GetUserId() (string, error) {
	path := "workspace.user-id"
	res := viper.GetString(path)
	if res == "" {
		return "", fmt.Errorf("workspace.user-id not found")
	}
	return res, nil
}

func (*workspace) SetUserId(userId string) error {
	viper.Set("workspace.user-id", userId)
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

func (*workspace) SetEmail(userId string) error {
	viper.Set("workspace.email", userId)
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
