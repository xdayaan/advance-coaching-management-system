package routes

import (
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupTeacherRoutes(router *gin.Engine, teacherHandler *handlers.TeacherHandler) {
	api := router.Group("/api")

	// Public routes (if any)
	// None for teachers - all require authentication

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())

	// Teacher profile routes (for teacher users)
	teacherProfile := protected.Group("/my-teacher-profile")
	teacherProfile.Use(middleware.RoleMiddleware("teacher"))
	{
		teacherProfile.GET("", teacherHandler.GetMyTeacherProfile)
		teacherProfile.PUT("", teacherHandler.UpdateMyTeacherProfile)
	}

	// Admin-only teacher management routes
	adminTeachers := protected.Group("/teachers")
	adminTeachers.Use(middleware.RoleMiddleware("admin"))
	{
		adminTeachers.POST("", teacherHandler.CreateTeacher)
		adminTeachers.GET("", teacherHandler.GetTeachers)
		adminTeachers.GET("/search", teacherHandler.SearchTeachers)
		adminTeachers.GET("/stats", teacherHandler.GetTeacherStats)
		adminTeachers.GET("/stats/salary", teacherHandler.GetSalaryStats)
		adminTeachers.GET("/stats/qualifications", teacherHandler.GetQualificationStats)
		adminTeachers.GET("/active", teacherHandler.GetActiveTeachers)
		adminTeachers.GET("/inactive", teacherHandler.GetInactiveTeachers)
		adminTeachers.POST("/bulk/status", teacherHandler.BulkUpdateTeacherStatus)
		adminTeachers.POST("/bulk/salary", teacherHandler.BulkUpdateSalary)
		adminTeachers.GET("/:id", teacherHandler.GetTeacher)
		adminTeachers.PUT("/:id", teacherHandler.UpdateTeacher)
		adminTeachers.DELETE("/:id", teacherHandler.DeleteTeacher)
		adminTeachers.PATCH("/:id/status", teacherHandler.ChangeTeacherStatus)
	}

	// Business-specific teacher routes (for business owners)
	businessTeachers := protected.Group("/businesses/:businessId/teachers")
	businessTeachers.Use(middleware.RoleMiddleware("admin", "business"))
	{
		businessTeachers.GET("", teacherHandler.GetTeachersByBusiness)
		businessTeachers.GET("/active", func(c *gin.Context) {
			// This would need a separate handler method or modify existing one
			// For now, redirect to general active teachers with business filter
			teacherHandler.GetActiveTeachers(c)
		})
		businessTeachers.GET("/inactive", func(c *gin.Context) {
			// Similar to above
			teacherHandler.GetInactiveTeachers(c)
		})
	}
}
