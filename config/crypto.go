package config

type crypto struct{}

var Crypto = crypto{}

func (*crypto) GetRecoveryPublicCertPath() (string, error) {
	return GetStringConfigOrPrintErr(
		"crypto.recovery-public-cert",
		"crypto.recovery-public-cert not found",
	)
}
