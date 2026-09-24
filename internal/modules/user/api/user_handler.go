package api

import (
	"strconv"

	"broadcast-platform/internal/modules/user/model"
	"broadcast-platform/internal/modules/user/service"
	"broadcast-platform/internal/platform/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type createUserRequest struct {
	Username        string `json:"username" binding:"required"`
	InitialPassword string `json:"initial_password" binding:"required"`
	Role            string `json:"role" binding:"required"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "username, initial_password, and role are required")
		return
	}
	safe, err := h.userService.CreateUser(c.Request.Context(), service.CreateUserRequest{
		Username:        req.Username,
		InitialPassword: req.InitialPassword,
		Role:            model.UserRole(req.Role),
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, safe)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	includeDeleted, _ := strconv.ParseBool(c.DefaultQuery("include_deleted", "false"))
	result, err := h.userService.ListUsers(c.Request.Context(), service.ListUsersRequest{
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: includeDeleted,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	safe, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, safe)
}

type updateUserRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body")
		return
	}
	safe, err := h.userService.UpdateUser(c.Request.Context(), id, service.UpdateUserRequest{
		Username: req.Username,
		Role:     model.UserRole(req.Role),
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, safe)
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "status is required")
		return
	}
	err := h.userService.UpdateUserStatus(c.Request.Context(), id, service.UpdateUserStatusRequest{
		Status: model.UserStatus(req.Status),
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.SuccessWithMessage(c, "status updated", nil)
}

type resetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "new_password is required")
		return
	}
	err := h.userService.ResetPassword(c.Request.Context(), id, service.ResetPasswordRequest{
		NewPassword: req.NewPassword,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.SuccessWithMessage(c, "password reset successfully", nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.SuccessWithMessage(c, "user deleted", nil)
}