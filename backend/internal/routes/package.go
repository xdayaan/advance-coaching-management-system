package routes

import (
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPackageRoutes(router *gin.RouterGroup, packageHandler *handlers.PackageHandler) {
	packages := router.Group("/packages")
	packages.Use(middleware.AuthMiddleware())

	// Public routes (for authenticated users)
	packages.GET("", packageHandler.GetPackages)
	packages.GET("/:id", packageHandler.GetPackage)
	packages.GET("/active", packageHandler.GetActivePackages)
	packages.GET("/search", packageHandler.SearchPackages)
	packages.GET("/price-range", packageHandler.GetPackagesByPriceRange)

	// Admin and Business only routes
	adminBusinessRoutes := packages.Group("")
	adminBusinessRoutes.Use(middleware.RoleMiddleware("admin", "business"))
	{
		adminBusinessRoutes.POST("", packageHandler.CreatePackage)
		adminBusinessRoutes.PUT("/:id", packageHandler.UpdatePackage)
		adminBusinessRoutes.DELETE("/:id", packageHandler.DeletePackage)
		adminBusinessRoutes.GET("/inactive", packageHandler.GetInactivePackages)
		adminBusinessRoutes.PATCH("/:id/status", packageHandler.ChangePackageStatus)
		adminBusinessRoutes.GET("/stats", packageHandler.GetPackageStats)
		adminBusinessRoutes.GET("/stats/prices", packageHandler.GetPriceStatistics)
		adminBusinessRoutes.PATCH("/bulk/status", packageHandler.BulkUpdatePackageStatus)
	}
}
