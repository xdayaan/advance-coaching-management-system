package routes

import (
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupStudentRoutes(router *gin.Engine, studentHandler *handlers.StudentHandler) {
	api := router.Group("/api")

	// Public routes (if any)
	// None for students - all require authentication

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())

	// Student profile routes (for student users)
	studentProfile := protected.Group("/my-student-profile")
	studentProfile.Use(middleware.RoleMiddleware("student"))
	{
		studentProfile.GET("", studentHandler.GetMyStudentProfile)
		studentProfile.PUT("", studentHandler.UpdateMyStudentProfile)
	}

	// Admin-only student management routes
	adminStudents := protected.Group("/students")
	adminStudents.Use(middleware.RoleMiddleware("admin"))
	{
		adminStudents.POST("", studentHandler.CreateStudent)
		adminStudents.GET("", studentHandler.GetStudents)
		adminStudents.GET("/search", studentHandler.SearchStudents)
		adminStudents.GET("/stats", studentHandler.GetStudentStats)
		adminStudents.GET("/stats/guardians", studentHandler.GetGuardianStats)
		adminStudents.GET("/active", studentHandler.GetActiveStudents)
		adminStudents.GET("/inactive", studentHandler.GetInactiveStudents)
		adminStudents.POST("/bulk/status", studentHandler.BulkUpdateStudentStatus)
		adminStudents.GET("/:id", studentHandler.GetStudent)
		adminStudents.PUT("/:id", studentHandler.UpdateStudent)
		adminStudents.DELETE("/:id", studentHandler.DeleteStudent)
		adminStudents.PATCH("/:id/status", studentHandler.ChangeStudentStatus)
	}

	// Business-specific student routes (for business owners)
	businessStudents := protected.Group("/businesses/:businessId/students")
	businessStudents.Use(middleware.RoleMiddleware("admin", "business"))
	{
		businessStudents.GET("", studentHandler.GetStudentsByBusiness)
		businessStudents.GET("/active", studentHandler.GetActiveStudentsByBusiness)
		businessStudents.GET("/inactive", studentHandler.GetInactiveStudentsByBusiness)
	}
}
