package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	config     *bootstrap.Config
	keyManager domainJWT.KeyManager
}

func NewJWTService(config *bootstrap.Config, keyManager domainJWT.KeyManager) *JWTService {
	service := &JWTService{
		config:     config,
		keyManager: keyManager,
	}
	err := keyManager.LoadKeys(config.Constants.JWTKeysPath.PrivateKey, config.Constants.JWTKeysPath.PublicKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to load JWT keys: %v", err))
	}

	return service
}

var _ usecase.TokenService = (*JWTService)(nil)

func (j *JWTService) GenerateTokens(ctx context.Context, userID uint64, isAdmin bool) (string, string, error) {

	accessTokenClaims := jwt.MapClaims{
		"sub":      userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(j.config.Env.JWT.AccessExpTime).Unix(),
		"iat":      time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(j.keyManager.GetPrivateKey())

	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshTokenClaims := jwt.MapClaims{
		"sub":      userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(j.config.Env.JWT.RefreshExpTime).Unix(),
		"iat":      time.Now().Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(j.keyManager.GetPrivateKey())

	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return accessTokenString, refreshTokenString, nil
}
