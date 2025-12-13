package service

import (
    "context"
    "crypto/rand"
    "crypto/rsa"
    "errors"
    "testing"
    "time"

    "github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
    "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
    "github.com/golang-jwt/jwt/v5"
    "github.com/stretchr/testify/assert"
)

func buildConfigForJWT() *bootstrap.Config {
    cfg := &bootstrap.Config{Constants: bootstrap.NewConstants(), Env: bootstrap.NewEnvironment()}
    // Short durations for tests
    cfg.Env.JWT.AccessExpTime = 5 * time.Minute
    cfg.Env.JWT.RefreshExpTime = 24 * time.Hour
    return cfg
}

func generateRSAKeys(t *testing.T) *rsa.PrivateKey {
    t.Helper()
    priv, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        t.Fatalf("failed to generate keys: %v", err)
    }
    return priv
}

func TestNewJWTService_LoadKeysPanicsOnError(t *testing.T) {
    cfg := buildConfigForJWT()
    km := mocks.NewMockKeyManager(t)
    // Expect LoadKeys to be called and return error
    km.On("LoadKeys", cfg.Constants.JWTKeysPath.PrivateKey, cfg.Constants.JWTKeysPath.PublicKey).Return(errors.New("load error"))
    assert.Panics(t, func() { _ = NewJWTService(cfg, km) })
    km.AssertExpectations(t)
}

func TestGenerateTokens_Success(t *testing.T) {
    cfg := buildConfigForJWT()
    km := mocks.NewMockKeyManager(t)
    // Generate real RSA keys
    priv := generateRSAKeys(t)
    pub := &priv.PublicKey

    // Expect LoadKeys called during constructor
    km.On("LoadKeys", cfg.Constants.JWTKeysPath.PrivateKey, cfg.Constants.JWTKeysPath.PublicKey).Return(nil)
    // Provide private key via mock; no need for GetPublicKey in service
    km.On("GetPrivateKey").Return(priv).Times(2)

    svc := NewJWTService(cfg, km)

    access, refresh, err := svc.GenerateTokens(context.Background(), 42, true)
    assert.NoError(t, err)
    assert.NotEmpty(t, access)
    assert.NotEmpty(t, refresh)

    // parse access token
    parsedA, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) { return pub, nil })
    assert.NoError(t, err)
    assert.True(t, parsedA.Valid)
    claimsA := parsedA.Claims.(jwt.MapClaims)
    assert.Equal(t, float64(42), claimsA["sub"]) // jwt numeric values as float64
    assert.Equal(t, true, claimsA["is_admin"])

    // parse refresh token
    parsedR, err := jwt.Parse(refresh, func(token *jwt.Token) (interface{}, error) { return pub, nil })
    assert.NoError(t, err)
    assert.True(t, parsedR.Valid)
    claimsR := parsedR.Claims.(jwt.MapClaims)
    assert.Equal(t, float64(42), claimsR["sub"]) // refresh contains same subject
    assert.Equal(t, true, claimsR["is_admin"])

    km.AssertExpectations(t)
}
