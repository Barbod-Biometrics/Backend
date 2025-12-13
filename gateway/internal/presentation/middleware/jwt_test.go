package middleware

import (
    "crypto/rand"
    "crypto/rsa"
    "fmt"
    "net/http/httptest"
    "testing"

    "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/stretchr/testify/assert"
)

func generateRSAKeys(t *testing.T) *rsa.PrivateKey {
    t.Helper()
    priv, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        t.Fatalf("failed to generate RSA key: %v", err)
    }
    return priv
}

func TestJWTMiddleware_MissingHeader(t *testing.T) {
    km := mocks.NewMockKeyManager(t)
    r := gin.New()
    r.Use(JWTMiddleware(km))
    r.GET("/", func(c *gin.Context) {})

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    r.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
}

func TestJWTMiddleware_InvalidHeaderFormat(t *testing.T) {
    km := mocks.NewMockKeyManager(t)
    r := gin.New()
    r.Use(JWTMiddleware(km))
    r.GET("/", func(c *gin.Context) {})

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "Bearer onlytoken extra")
    r.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
}

func TestJWTMiddleware_InvalidSignature(t *testing.T) {
    // sign token with key A, but middleware's GetPublicKey returns key B
    privA := generateRSAKeys(t)
    privB := generateRSAKeys(t)
    pubB := &privB.PublicKey

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": float64(123), "is_admin": true})
    tokenStr, err := token.SignedString(privA)
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }

    km := mocks.NewMockKeyManager(t)
    km.On("GetPublicKey").Return(pubB)

    r := gin.New()
    r.Use(JWTMiddleware(km))
    r.GET("/", func(c *gin.Context) {})

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
    r.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
    km.AssertExpectations(t)
}

func TestJWTMiddleware_ValidToken_FloatSub(t *testing.T) {
    priv := generateRSAKeys(t)
    pub := &priv.PublicKey

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": float64(999), "is_admin": true})
    tokenStr, err := token.SignedString(priv)
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }

    km := mocks.NewMockKeyManager(t)
    km.On("GetPublicKey").Return(pub)

    r := gin.New()
    r.Use(JWTMiddleware(km))
    var called bool
    r.GET("/", func(c *gin.Context) {
        called = true
        // verify context values
        uid, err := GetUserIDFromContext(c)
        if err != nil {
            t.Fatalf("failed to get user id: %v", err)
        }
        if uid != 999 {
            t.Fatalf("unexpected user id: %d", uid)
        }
        if !IsAdminFromContext(c) {
            t.Fatalf("expected admin true")
        }
    })

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
    r.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.True(t, called)
    km.AssertExpectations(t)
}

func TestJWTMiddleware_ValidToken_SubString(t *testing.T) {
    priv := generateRSAKeys(t)
    pub := &priv.PublicKey

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "42", "is_admin": false})
    tokenStr, err := token.SignedString(priv)
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }

    km := mocks.NewMockKeyManager(t)
    km.On("GetPublicKey").Return(pub)

    r := gin.New()
    r.Use(JWTMiddleware(km))
    var called bool
    r.GET("/", func(c *gin.Context) {
        called = true
        uid, err := GetUserIDFromContext(c)
        if err != nil {
            t.Fatalf("failed to get user id: %v", err)
        }
        if uid != 42 {
            t.Fatalf("unexpected user id: %d", uid)
        }
        if IsAdminFromContext(c) {
            t.Fatalf("expected admin false")
        }
    })

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
    r.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.True(t, called)
    km.AssertExpectations(t)
}

func TestJWTMiddleware_InvalidClaims(t *testing.T) {
    // Create token with sub as array (unsupported) and ensure middleware rejects
    priv := generateRSAKeys(t)
    pub := &priv.PublicKey

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": []int{1, 2, 3}, "is_admin": true})
    tokenStr, err := token.SignedString(priv)
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }

    km := mocks.NewMockKeyManager(t)
    km.On("GetPublicKey").Return(pub)

    r := gin.New()
    r.Use(JWTMiddleware(km))
    r.GET("/", func(c *gin.Context) {})

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
    r.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
    km.AssertExpectations(t)
}

func TestExtractUserIDVariousTypes(t *testing.T) {
    var v uint64
    var err error
    v, err = extractUserID(float64(7))
    assert.NoError(t, err)
    assert.Equal(t, uint64(7), v)

    v, err = extractUserID("8")
    assert.NoError(t, err)
    assert.Equal(t, uint64(8), v)

    v, err = extractUserID(int(9))
    assert.NoError(t, err)
    assert.Equal(t, uint64(9), v)

    v, err = extractUserID(int64(10))
    assert.NoError(t, err)
    assert.Equal(t, uint64(10), v)

    v, err = extractUserID(uint64(11))
    assert.NoError(t, err)
    assert.Equal(t, uint64(11), v)

    _, err = extractUserID([]int{1})
    assert.Error(t, err)
}

func TestExtractIsAdmin(t *testing.T) {
    assert.False(t, extractIsAdmin(nil))
    assert.True(t, extractIsAdmin(true))
    assert.False(t, extractIsAdmin("notbool"))
}

func TestGetUserIDFromContextAndAdmin(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Set(ContextKeyUserID, uint64(55))
    c.Set(ContextKeyIsAdmin, true)
    uid, err := GetUserIDFromContext(c)
    assert.NoError(t, err)
    assert.Equal(t, uint64(55), uid)
    assert.True(t, IsAdminFromContext(c))

    // invalid types
    c2, _ := gin.CreateTestContext(w)
    c2.Set(ContextKeyUserID, "notuint")
    _, err = GetUserIDFromContext(c2)
    assert.Error(t, err)
    c2.Set(ContextKeyUserID, uint64(55))
    c2.Set(ContextKeyIsAdmin, "notbool")
    assert.False(t, IsAdminFromContext(c2))
}

func TestAdminMiddleware(t *testing.T) {
    r := gin.New()
    r.Use(AdminMiddleware())
    r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

    // user not admin
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    // set context via middleware; inject a middleware to set isAdmin false
    r2 := gin.New()
    r2.Use(func(c *gin.Context) { c.Set(ContextKeyIsAdmin, false); c.Next() })
    r2.Use(AdminMiddleware())
    r2.GET("/", func(c *gin.Context) { c.String(200, "ok") })
    r2.ServeHTTP(w, req)
    assert.Equal(t, 403, w.Code)

    // user is admin
    w2 := httptest.NewRecorder()
    r3 := gin.New()
    r3.Use(func(c *gin.Context) { c.Set(ContextKeyIsAdmin, true); c.Next() })
    r3.Use(AdminMiddleware())
    r3.GET("/", func(c *gin.Context) { c.String(200, "ok") })
    r3.ServeHTTP(w2, req)
    assert.Equal(t, 200, w2.Code)
}
