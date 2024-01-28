package config

import "github.com/spf13/viper"

type crypto struct{}

var Crypto = crypto{}

func (*crypto) GetRecoveryPublicCertPath() (string, error) {
	return GetStringConfigOrPrintErr(
		"crypto.recovery-public-cert",
		"crypto.recovery-public-cert not found",
	)
}

func (*crypto) GetChunkSize() (uint64, error) {
	return viper.GetUint64("crypto.chunk-size"), nil
}
