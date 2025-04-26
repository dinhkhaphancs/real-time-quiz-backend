package middleware

import (
	"net/http"
	"strings"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/auth"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// AuthUserKey is the key used to store authenticated user in the context
	AuthUserKey = "auth_user"
	// AuthorizationHeaderKey is the key for authorization header
	AuthorizationHeaderKey = "Authorization"
	// BearerToken is the prefix for token-based authentication
	BearerToken = "Bearer"
)

// JWTAuthMiddleware creates a middleware for JWT authentication
func JWTAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			response.WithError(c, http.StatusUnauthorized, "Unauthorized", "Authorization header is required")
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || fields[0] != BearerToken {
			response.WithError(c, http.StatusUnauthorized, "Unauthorized", "Invalid authorization format. Format should be 'Bearer {token}'")
			c.Abort()
			return
		}

		// Extract the token
		tokenString := fields[1]

		// Validate the token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			var statusCode int
			var message string

			if err == auth.ErrExpiredToken {
				statusCode = http.StatusUnauthorized
				message = "Token has expired"
			} else {
				statusCode = http.StatusUnauthorized
				message = "Invalid token"
			}

			response.WithError(c, statusCode, "Unauthorized", message)
			c.Abort()
			return
		}

		// Create a user object from claims
		user := &model.User{
			ID:    claims.UserID,
			Email: claims.Email,
		}

		// Set user in context
		c.Set(AuthUserKey, user)
		c.Next()
	}
}

// GetAuthUser retrieves the authenticated user from the context
func GetAuthUser(c *gin.Context) *model.User {
	user, exists := c.Get(AuthUserKey)
	if !exists {
		return nil
	}
	return user.(*model.User)
}

// GetAuthUserID retrieves the authenticated user ID from the context
func GetAuthUserID(c *gin.Context) uuid.UUID {
	user := GetAuthUser(c)
	if user == nil {
		return uuid.Nil
	}
	return user.ID
}
