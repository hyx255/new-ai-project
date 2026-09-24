package api

import (
	"errors"
	"net/http"
	"strconv"

	"broadcast-platform/internal/modules/device/service"
	"broadcast-platform/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// DeviceHandler handles HTTP requests for Device operations.
type DeviceHandler struct {
	service *service.DeviceService
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{service: svc}
}

// CreateDeviceRequest represents the request body for creating a Device.
type CreateDeviceRequest struct {
	Name         string `json:"name" binding:"required"`
	DeviceTypeID string `json:"device_type_id" binding:"required"`
	Address      string `json:"address"`
}

// UpdateDeviceRequest represents the request body for updating a Device.
type UpdateDeviceRequest struct {
	Name         string `json:"name" binding:"required"`
	DeviceTypeID string `json:"device_type_id"`
	Address      string `json:"address"`
}

// UpdateDeviceStatusRequest represents the request body for updating Device status.
type UpdateDeviceStatusRequest struct {
	Action string `json:"action" binding:"required"` // "enable" or "disable"
}

// Create handles POST /api/devices
func (h *DeviceHandler) Create(c *gin.Context) {
	var req CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	params := service.CreateDeviceParams{
		Name:         req.Name,
		DeviceTypeID: req.DeviceTypeID,
		Address:      req.Address,
	}

	device, err := h.service.Create(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrDeviceTypeInvalid) {
			response.NotFound(c, "device type not found")
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, device)
}

// GetByID handles GET /api/devices/:id
func (h *DeviceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	device, dt, err := h.service.GetByIDWithDeviceType(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, "device not found")
			return
		}
		response.Error(c, err)
		return
	}

	// Build response with DeviceType info and capabilities
	data := gin.H{
		"id":             device.ID,
		"name":           device.Name,
		"device_type_id": device.DeviceTypeID,
		"address":        device.Address,
		"status":         device.Status,
		"last_online_at": device.LastOnlineAt,
		"created_at":     device.CreatedAt,
		"updated_at":     device.UpdatedAt,
	}

	if dt != nil {
		data["device_type"] = gin.H{
			"id":           dt.ID,
			"name":         dt.Name,
			"vendor":       dt.Vendor,
			"model":        dt.Model,
			"capabilities": dt.Capabilities,
		}
	}

	response.Success(c, data)
}

// List handles GET /api/devices
func (h *DeviceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	result, err := h.service.List(c.Request.Context(), page, pageSize, search)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":     result.Items,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	})
}

// Update handles PUT /api/devices/:id
func (h *DeviceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	params := service.UpdateDeviceParams{
		Name:         req.Name,
		DeviceTypeID: req.DeviceTypeID,
		Address:      req.Address,
	}

	device, err := h.service.Update(c.Request.Context(), id, params)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, "device not found")
			return
		}
		if errors.Is(err, service.ErrDeviceTypeInvalid) {
			response.NotFound(c, "device type not found")
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, device)
}

// UpdateStatus handles PATCH /api/devices/:id/status
func (h *DeviceHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	var device interface{}
	var err error

	switch req.Action {
	case "disable":
		device, err = h.service.Disable(c.Request.Context(), id)
	case "enable":
		device, err = h.service.Enable(c.Request.Context(), id)
	default:
		response.ValidationError(c, "action must be 'enable' or 'disable'")
		return
	}

	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, "device not found")
			return
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    2101,
				"message": err.Error(),
			})
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, device)
}
