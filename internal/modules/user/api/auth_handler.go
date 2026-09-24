package api

import (
	"net/http"

	"broadcast-platform/internal/modules/user/service"
	"broadcast-platform/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest is the JSON body for POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "username and password are required")
		return
	}

	loginResp, err := h.authService.Login(c.Request.Context(), service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Response{
			Code:    3001,
			Message: "invalid credentials",
		})
		return
	}

	response.Success(c, gin.H{
		"access_token": loginResp.AccessToken,
		"user":         loginResp.User,
	})
}

// GetMe handles GET /api/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	user, err := h.authService.GetCurrentUser(c.Request.Context())
	if err != nil {
		response.Unauthorized(c, "unable to load current user")
		return
	}

	safe := user.ToSafe()
	response.Success(c, safe)
}

// ChangePasswordRequest is the JSON body for POST /api/auth/change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// ChangePassword handles POST /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "current_password and new_password are required")
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), service.ChangePasswordRequest{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "password changed successfully", nil)
}

// handleServiceError maps service-layer errors to appropriate HTTP responses.
func handleServiceError(c *gin.Context, err error) {
	msg := err.Error()

	switch msg {
	case "current password is incorrect":
		c.JSON(http.StatusBadRequest, response.Response{
			Code:    3001,
			Message: "current password is incorrect",
		})
	case "new password must be different from current password":
		response.ValidationError(c, "new password must be different from current password")
	case "username already exists":
		c.JSON(http.StatusConflict, response.Response{
			Code:    3101,
			Message: "username already exists",
		})
	case "user not found":
		response.NotFound(c, "user not found")
	case "forbidden: admin role required":
		response.Forbidden(c, "admin access required")
	case "forbidden: cannot perform this action on yourself":
		response.Forbidden(c, "cannot perform this action on yourself")
	case "invalid role: must be ADMIN or USER":
		response.ValidationError(c, "invalid role")
	case "username is required":
		response.ValidationError(c, "username is required")
	case "invalid status: must be ACTIVE or DISABLED":
		response.ValidationError(c, "invalid status")
	default:
		if msg == "user is already ACTIVE" || msg == "user is already DISABLED" {
			response.ValidationError(c, msg)
			return
		}
		if isPasswordPolicyError(msg) {
			response.ValidationError(c, msg)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Response{
			Code:    1999,
			Message: "internal server error",
		})
	}
}

func isPasswordPolicyError(msg string) bool {
	policyErrors := []string{
		"password must be at least 8 characters",
		"password must contain at least one uppercase letter",
		"password must contain at least one lowercase letter",
		"password must contain at least one digit",
	}
	for _, pe := range policyErrors {
		if msg == pe {
			return true
		}
	}
	return false
}
