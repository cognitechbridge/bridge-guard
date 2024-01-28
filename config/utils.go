package config

import (
	"fmt"
	"github.com/spf13/viper"
)

func GetStringConfigOrPrintErr(path string, err string) (string, error) {
	res := viper.GetString(path)
	if res == "" {
		fmt.Println(err)
		return "", fmt.Errorf(err)
	}
	return res, nil
}
