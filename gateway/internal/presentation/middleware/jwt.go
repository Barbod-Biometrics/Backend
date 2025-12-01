package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKeyUserID is the key used to store user ID in gin context
const ContextKeyUserID = "userID"

// JWTMiddleware creates a middleware that validates JWT tokens and extracts user ID
func JWTMiddleware(keyManager domainJWT.KeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(tok *jwt.Token) (interface{}, error) {
			// Ensure signing method is RSA
			if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
			}
			return keyManager.GetPublicKey(), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// Extract user ID from "sub" claim
		userID, err := extractUserID(claims["sub"])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}

		// Set user ID in context for handlers to use
		c.Set(ContextKeyUserID, userID)
		c.Next()
	}
}

// extractUserID extracts user ID from the sub claim which can be float64, string, or int
func extractUserID(sub interface{}) (uint64, error) {
	switch v := sub.(type) {
	case float64:
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	case int:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint64:
		return v, nil
	default:
		return 0, fmt.Errorf("unexpected sub claim type: %T", sub)
	}
}

// GetUserIDFromContext extracts user ID from gin context
// Returns 0 if user ID is not found or invalid
func GetUserIDFromContext(c *gin.Context) uint64 {
	v, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0
	}
	if uid, ok := v.(uint64); ok {
		return uid
	}
	return 0
}
