package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/service"
	"broadcast-platform/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// OperationHandler handles HTTP requests for Operation operations.
type OperationHandler struct {
	service *service.OperationService
}

// NewOperationHandler creates a new OperationHandler.
func NewOperationHandler(svc *service.OperationService) *OperationHandler {
	return &OperationHandler{service: svc}
}

// CreateOperationRequest represents the request body for creating an Operation.
type CreateOperationRequest struct {
	DeviceID   string          `json:"device_id" binding:"required"`
	Type       string          `json:"type" binding:"required"`
	Parameters json.RawMessage `json:"parameters"`
}

// Create handles POST /api/operations
func (h *OperationHandler) Create(c *gin.Context) {
	var req CreateOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	opType := model.OperationType(req.Type)
	if !opType.IsValid() {
		response.ValidationError(c, "invalid operation type")
		return
	}

	params := service.CreateOperationParams{
		DeviceID:   req.DeviceID,
		Type:       opType,
		Parameters: req.Parameters,
	}

	result, err := h.service.Create(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			response.NotFound(c, "device not found")
			return
		}
		if errors.Is(err, service.ErrDeviceTypeInvalid) {
			response.NotFound(c, "device type not found")
			return
		}
		if errors.Is(err, service.ErrDeviceDisabled) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    3001,
				"message": "device is disabled",
			})
			return
		}
		if errors.Is(err, service.ErrCapabilityRequired) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    3002,
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrInvalidParameter) {
			response.ValidationError(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrConcurrentOperation) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    3003,
				"message": "device has an active operation",
			})
			return
		}
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"operation": result.Operation,
			"execution": result.Execution,
		},
	})
}

// GetByID handles GET /api/operations/:id
func (h *OperationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	op, execs, err := h.service.GetOperationWithExecutions(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrOperationNotFound) {
			response.NotFound(c, "operation not found")
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"operation":  op,
		"executions": execs,
	})
}

// ListByDevice handles GET /api/devices/:id/operations
func (h *OperationHandler) ListByDevice(c *gin.Context) {
	deviceID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	operations, total, err := h.service.ListDeviceOperations(c.Request.Context(), deviceID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	if operations == nil {
		operations = []*model.Operation{}
	}

	response.Success(c, gin.H{
		"items":     operations,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}