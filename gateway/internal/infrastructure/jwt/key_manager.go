package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"
	"sync"

	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/golang-jwt/jwt/v5"
)

type JWTKeyManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	mutex      sync.RWMutex
	isLoaded   bool
}

func NewJWTKeyManager() *JWTKeyManager {
	return &JWTKeyManager{}
}

var _ domainJWT.KeyManager = (*JWTKeyManager)(nil)

func (k *JWTKeyManager) LoadKeys(privateKeyPath, publicKeyPath string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	privKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	k.privateKey = privateKey
	k.publicKey = publicKey
	k.isLoaded = true

	return nil
}

func (k *JWTKeyManager) GetPrivateKey() *rsa.PrivateKey {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	if !k.isLoaded {
		return nil
	}
	return k.privateKey
}

func (k *JWTKeyManager) GetPublicKey() *rsa.PublicKey {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	if !k.isLoaded {
		return nil
	}
	return k.publicKey
}
