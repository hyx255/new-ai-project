package api

import (
	"net/http"
	"strings"

	"broadcast-platform/internal/modules/user/model"
	"broadcast-platform/internal/modules/user/repository"
	"broadcast-platform/internal/modules/user/service"
	"broadcast-platform/internal/platform/auth"
	"broadcast-platform/internal/platform/errors"
	"broadcast-platform/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and loads the current user into context.
func AuthMiddleware(tokenProvider *auth.TokenProvider, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip public routes
		path := c.Request.URL.Path
		if isPublicRoute(path) {
			c.Next()
			return
		}

		// Extract Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			response.Unauthorized(c, "token required")
			c.Abort()
			return
		}

		// Parse and validate JWT
		claims, err := tokenProvider.Parse(tokenString)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				c.JSON(http.StatusUnauthorized, response.Response{
					Code:    errors.CodeTokenExpired,
					Message: "token expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, response.Response{
					Code:    errors.CodeTokenInvalid,
					Message: "invalid token",
				})
			}
			c.Abort()
			return
		}

		// Load user from database to check current state
		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code:    errors.CodeTokenInvalid,
				Message: "invalid token",
			})
			c.Abort()
			return
		}

		// Check if user is deleted
		if user.IsDeleted() {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code:    errors.CodeAccountDeleted,
				Message: "account has been deleted",
			})
			c.Abort()
			return
		}

		// Check if user is disabled
		if user.Status != model.StatusActive {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code:    errors.CodeAccountDisabled,
				Message: "account is disabled",
			})
			c.Abort()
			return
		}

		// Verify token_version matches (ensures old tokens are invalidated)
		if claims.TokenVersion != user.TokenVersion {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code:    errors.CodeTokenInvalid,
				Message: "token has been invalidated",
			})
			c.Abort()
			return
		}

		// Store current user in context
		currentUser := &service.CurrentUser{
			ID:                 user.ID,
			Username:           user.Username,
			Role:               user.Role,
			MustChangePassword: user.MustChangePassword,
			TokenVersion:       user.TokenVersion,
		}
		ctx := service.WithCurrentUser(c.Request.Context(), currentUser)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequirePasswordChangedMiddleware blocks business API access if must_change_password is true.
// Only allows GET /api/auth/me and POST /api/auth/change-password.
func RequirePasswordChangedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip public routes
		if isPublicRoute(path) {
			c.Next()
			return
		}

		// Allow auth endpoints even if must_change_password
		if isPasswordChangeAllowedRoute(path) {
			c.Next()
			return
		}

		// Check if user must change password
		currentUser := service.GetCurrentUser(c.Request.Context())
		if currentUser != nil && currentUser.MustChangePassword {
			c.JSON(http.StatusForbidden, response.Response{
				Code:    errors.CodeMustChangePassword,
				Message: "password change required before accessing this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminMiddleware ensures only ADMIN users can access certain routes.
func RequireAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := service.RequireAdmin(c.Request.Context()); err != nil {
			c.JSON(http.StatusForbidden, response.Response{
				Code:    errors.CodeForbidden,
				Message: "admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// isPublicRoute returns true for routes that don't require authentication.
func isPublicRoute(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/auth/login",
		"/metrics",
	}
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	// Allow debug/pprof routes
	if strings.HasPrefix(path, "/debug/") {
		return true
	}
	return false
}

// isPasswordChangeAllowedRoute returns true for routes accessible even when must_change_password is true.
func isPasswordChangeAllowedRoute(path string) bool {
	allowed := []string{
		"/api/auth/me",
		"/api/auth/change-password",
	}
	for _, p := range allowed {
		if path == p {
			return true
		}
	}
	return false
}
