package api

import "github.com/gin-gonic/gin"

// RegisterRoutes registers User module API routes.
// Auth routes are public; User routes require ADMIN role.
func RegisterRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	userHandler *UserHandler,
) {
	api := router.Group("/api")

	// Auth routes (public or authenticated)
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", authHandler.GetMe)
		auth.POST("/change-password", authHandler.ChangePassword)
	}

	// User management routes (ADMIN only)
	users := api.Group("/users")
	users.Use(RequireAdminMiddleware())
	{
		users.POST("", userHandler.CreateUser)
		users.GET("", userHandler.ListUsers)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
		users.PATCH("/:id/status", userHandler.UpdateStatus)
		users.POST("/:id/reset-password", userHandler.ResetPassword)
	}
}
