package api

import "github.com/gin-gonic/gin"

// RegisterRoutes registers Device module API routes.
func RegisterRoutes(
	router *gin.Engine,
	dtHandler *DeviceTypeHandler,
	devHandler *DeviceHandler,
	opHandler *OperationHandler,
	batchHandler *BatchOperationHandler,
) {
	api := router.Group("/api")

	// DeviceType routes
	deviceTypes := api.Group("/device-types")
	{
		deviceTypes.POST("", dtHandler.Create)
		deviceTypes.GET("", dtHandler.List)
		deviceTypes.GET("/:id", dtHandler.GetByID)
		deviceTypes.PUT("/:id", dtHandler.Update)
		deviceTypes.DELETE("/:id", dtHandler.Delete)
	}

	// Device routes
	devices := api.Group("/devices")
	{
		devices.POST("", devHandler.Create)
		devices.GET("", devHandler.List)
		devices.GET("/:id", devHandler.GetByID)
		devices.PUT("/:id", devHandler.Update)
		devices.PATCH("/:id/status", devHandler.UpdateStatus)
		// Device operation history
		devices.GET("/:id/operations", opHandler.ListByDevice)
	}

	// Single-device operation routes (backward compatible)
	operations := api.Group("/operations")
	{
		operations.POST("", opHandler.Create)
		operations.GET("/:id", opHandler.GetByID)
	}

	// Batch operation routes
	batchOps := api.Group("/batch-operations")
	{
		batchOps.POST("", batchHandler.Create)
		batchOps.GET("/:id", batchHandler.GetByID)
		batchOps.GET("/:id/progress", batchHandler.GetProgress)
	}
}