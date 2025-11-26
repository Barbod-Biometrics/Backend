package jwt

import "crypto/rsa"

type KeyManager interface {
	LoadKeys(privateKeyPath, publicKeyPath string) error
	GetPrivateKey() *rsa.PrivateKey
	GetPublicKey() *rsa.PublicKey
}
