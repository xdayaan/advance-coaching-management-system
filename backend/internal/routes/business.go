package routes

import (
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupBusinessRoutes(router *gin.RouterGroup, businessHandler *handlers.BusinessHandler) {
	// Public route - get business by slug (no auth required)
	router.GET("/business/:slug", businessHandler.GetBusinessBySlug)

	// Business profile routes (for business users)
	businessProfile := router.Group("/my-business")
	businessProfile.Use(middleware.AuthMiddleware())
	businessProfile.Use(middleware.RoleMiddleware("business"))
	{
		businessProfile.GET("", businessHandler.GetMyBusiness)
		businessProfile.PUT("", businessHandler.UpdateMyBusiness)
	}

	// Admin business management routes
	businesses := router.Group("/businesses")
	businesses.Use(middleware.AuthMiddleware())
	businesses.Use(middleware.RoleMiddleware("admin"))
	{
		// Essential CRUD operations
		businesses.POST("", businessHandler.CreateBusiness)
		businesses.GET("", businessHandler.GetBusinesses)
		businesses.GET("/:id", businessHandler.GetBusiness)
		businesses.PUT("/:id", businessHandler.UpdateBusiness)
		businesses.DELETE("/:id", businessHandler.DeleteBusiness)

		// Status management
		businesses.PATCH("/:id/status", businessHandler.ChangeBusinessStatus)
		businesses.GET("/active", businessHandler.GetActiveBusinesses)
		businesses.GET("/inactive", businessHandler.GetInactiveBusinesses)

		// Package management
		businesses.POST("/:id/assign-package", businessHandler.AssignPackage)
		businesses.DELETE("/:id/remove-package", businessHandler.RemovePackage)
		businesses.GET("/package/:packageId", businessHandler.GetBusinessesByPackage)
		businesses.GET("/no-package", businessHandler.GetBusinessesWithoutPackage)

		// Search functionality
		businesses.GET("/search", businessHandler.SearchBusinesses)

		// Location functionality
		businesses.GET("/by-location", businessHandler.GetBusinessesByLocation)
		businesses.GET("/locations", businessHandler.GetBusinessLocations)

		// Statistics and reporting
		businesses.GET("/stats", businessHandler.GetBusinessStats)
		businesses.GET("/stats/locations", businessHandler.GetLocationStats)
		businesses.GET("/stats/packages", businessHandler.GetPackageDistribution)

		// Bulk operations
		businesses.POST("/bulk/status", businessHandler.BulkUpdateStatus)
		businesses.POST("/bulk/assign-package", businessHandler.BulkAssignPackage)
	}
}
