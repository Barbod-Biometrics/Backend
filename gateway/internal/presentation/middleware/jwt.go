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

// Context keys for storing user information
const (
	ContextKeyUserID  = "userID"
	ContextKeyIsAdmin = "isAdmin"
)

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

			return keyManager.GetPublicKey(), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
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

		// Extract is_admin claim (defaults to false if not present)
		isAdmin := extractIsAdmin(claims["is_admin"])

		// Set user ID and admin status in context for handlers to use
		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyIsAdmin, isAdmin)
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

// extractIsAdmin extracts is_admin from the claim (defaults to false)
func extractIsAdmin(claim interface{}) bool {
	if claim == nil {
		return false
	}
	if isAdmin, ok := claim.(bool); ok {
		return isAdmin
	}
	return false
}

// GetUserIDFromContext extracts user ID from gin context.
// Returns the user ID and an error if not found or invalid.
func GetUserIDFromContext(c *gin.Context) (uint64, error) {
	v, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, fmt.Errorf("user ID not found in context")
	}
	if uid, ok := v.(uint64); ok {
		return uid, nil
	}
	return 0, fmt.Errorf("user ID in context has invalid type: %T", v)
}

// IsAdminFromContext checks if the current user is an admin.
// Returns false if not found or not an admin.
func IsAdminFromContext(c *gin.Context) bool {
	v, exists := c.Get(ContextKeyIsAdmin)
	if !exists {
		return false
	}
	if isAdmin, ok := v.(bool); ok {
		return isAdmin
	}
	return false
}

// AdminMiddleware is a middleware that requires the user to be an admin.
// It must be used AFTER JWTMiddleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdminFromContext(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
