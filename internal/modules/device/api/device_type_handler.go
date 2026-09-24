package api

import (
	"errors"
	"net/http"

	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/service"
	"broadcast-platform/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// DeviceTypeHandler handles HTTP requests for DeviceType operations.
type DeviceTypeHandler struct {
	service *service.DeviceTypeService
}

// NewDeviceTypeHandler creates a new DeviceTypeHandler.
func NewDeviceTypeHandler(svc *service.DeviceTypeService) *DeviceTypeHandler {
	return &DeviceTypeHandler{service: svc}
}

// CreateDeviceTypeRequest represents the request body for creating a DeviceType.
type CreateDeviceTypeRequest struct {
	Name         string             `json:"name" binding:"required"`
	Vendor       string             `json:"vendor" binding:"required"`
	Model        string             `json:"model" binding:"required"`
	Description  string             `json:"description"`
	Capabilities []model.Capability `json:"capabilities" binding:"required"`
}

// UpdateDeviceTypeRequest represents the request body for updating a DeviceType.
type UpdateDeviceTypeRequest struct {
	Name         string             `json:"name" binding:"required"`
	Vendor       string             `json:"vendor" binding:"required"`
	Model        string             `json:"model" binding:"required"`
	Description  string             `json:"description"`
	Capabilities []model.Capability `json:"capabilities" binding:"required"`
}

// Create handles POST /api/device-types
func (h *DeviceTypeHandler) Create(c *gin.Context) {
	var req CreateDeviceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	params := service.CreateDeviceTypeParams{
		Name:         req.Name,
		Vendor:       req.Vendor,
		Model:        req.Model,
		Description:  req.Description,
		Capabilities: req.Capabilities,
	}

	dt, err := h.service.Create(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrDeviceTypeAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    2001,
				"message": err.Error(),
			})
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, dt)
}

// GetByID handles GET /api/device-types/:id
func (h *DeviceTypeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	dt, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDeviceTypeNotFound) {
			response.NotFound(c, "device type not found")
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, dt)
}

// List handles GET /api/device-types
func (h *DeviceTypeHandler) List(c *gin.Context) {
	dts, err := h.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	if dts == nil {
		dts = []*model.DeviceType{}
	}

	response.Success(c, gin.H{
		"items": dts,
		"total": len(dts),
	})
}

// Update handles PUT /api/device-types/:id
func (h *DeviceTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDeviceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	params := service.UpdateDeviceTypeParams{
		Name:         req.Name,
		Vendor:       req.Vendor,
		Model:        req.Model,
		Description:  req.Description,
		Capabilities: req.Capabilities,
	}

	dt, err := h.service.Update(c.Request.Context(), id, params)
	if err != nil {
		if errors.Is(err, service.ErrDeviceTypeNotFound) {
			response.NotFound(c, "device type not found")
			return
		}
		if errors.Is(err, service.ErrDeviceTypeAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    2001,
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrCannotRemoveCapability) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    2002,
				"message": err.Error(),
			})
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, dt)
}

// Delete handles DELETE /api/device-types/:id
func (h *DeviceTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDeviceTypeNotFound) {
			response.NotFound(c, "device type not found")
			return
		}
		if errors.Is(err, service.ErrDeviceTypeInUse) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    2003,
				"message": err.Error(),
			})
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "device type deleted",
	})
}
