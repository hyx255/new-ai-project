package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/service"
	"broadcast-platform/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// BatchOperationHandler handles HTTP requests for batch operations.
type BatchOperationHandler struct {
	service *service.BatchOperationService
}

// NewBatchOperationHandler creates a new BatchOperationHandler.
func NewBatchOperationHandler(svc *service.BatchOperationService) *BatchOperationHandler {
	return &BatchOperationHandler{service: svc}
}

// CreateBatchRequest represents the request body for creating a batch operation.
type CreateBatchRequest struct {
	DeviceIDs  []string        `json:"device_ids" binding:"required"`
	Type       string          `json:"type" binding:"required"`
	Parameters json.RawMessage `json:"parameters"`
}

// Create handles POST /api/batch-operations
func (h *BatchOperationHandler) Create(c *gin.Context) {
	var req CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	opType := model.OperationType(req.Type)
	if !opType.IsValid() {
		response.ValidationError(c, "invalid operation type")
		return
	}

	params := service.CreateBatchParams{
		DeviceIDs:  req.DeviceIDs,
		Type:       opType,
		Parameters: req.Parameters,
	}

	result, err := h.service.CreateBatchOperation(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrEmptyDeviceIDs) {
			response.ValidationError(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrBatchTooLarge) {
			response.ValidationError(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrDuplicateDeviceID) {
			response.ValidationError(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrInvalidParameter) {
			response.ValidationError(c, err.Error())
			return
		}
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"operation":  result.Operation,
			"executions": result.Executions,
		},
	})
}

// GetByID handles GET /api/batch-operations/:id
func (h *BatchOperationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	op, execs, err := h.service.GetBatchOperation(c.Request.Context(), id)
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

// GetProgress handles GET /api/batch-operations/:id/progress
func (h *BatchOperationHandler) GetProgress(c *gin.Context) {
	id := c.Param("id")

	progress, err := h.service.GetProgress(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrOperationNotFound) {
			response.NotFound(c, "operation not found")
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, progress)
}