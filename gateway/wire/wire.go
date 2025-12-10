package wire

import "github.com/Barbod-Biometrics/Backend/gateway/bootstrap"

func ProvideConfig() *bootstrap.Config {
	return bootstrap.Run()
}
