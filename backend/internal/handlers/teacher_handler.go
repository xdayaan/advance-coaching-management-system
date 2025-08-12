package handlers

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TeacherHandler struct {
	teacherService services.TeacherService
}

func NewTeacherHandler(teacherService services.TeacherService) *TeacherHandler {
	return &TeacherHandler{
		teacherService: teacherService,
	}
}

// CreateTeacher godoc
// @Summary Create a new teacher
// @Description Create a new teacher (Admin/Business only)
// @Tags teachers
// @Accept json
// @Produce json
// @Param request body models.CreateTeacherRequest true "Teacher data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Success response with teacher data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/teachers [post]
func (h *TeacherHandler) CreateTeacher(c *gin.Context) {
	var req models.CreateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	teacher, err := h.teacherService.CreateTeacher(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Teacher created successfully",
		"data":    teacher,
	})
}

// GetTeachers godoc
// @Summary Get all teachers
// @Description Get all teachers with pagination and filters (Admin only)
// @Tags teachers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query int false "Filter by status (0=inactive, 1=active)"
// @Param business_id query int false "Filter by business ID"
// @Param min_salary query number false "Filter by minimum salary"
// @Param max_salary query number false "Filter by maximum salary"
// @Param qualification query string false "Filter by qualification"
// @Param search query string false "Search in name, qualification, or experience"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with teachers list"
// @Router /api/teachers [get]
func (h *TeacherHandler) GetTeachers(c *gin.Context) {
	var filters repository.TeacherFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	teachers, total, err := h.teacherService.GetTeachers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get teachers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"teachers": teachers,
			"total":    total,
			"page":     filters.Page,
			"limit":    filters.Limit,
		},
	})
}

// GetTeacher godoc
// @Summary Get teacher by ID
// @Description Get a specific teacher by ID
// @Tags teachers
// @Accept json
// @Produce json
// @Param id path int true "Teacher ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with teacher data"
// @Failure 404 {object} map[string]string "Teacher not found"
// @Router /api/teachers/{id} [get]
func (h *TeacherHandler) GetTeacher(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid teacher ID",
		})
		return
	}

	teacher, err := h.teacherService.GetTeacherByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Teacher not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    teacher,
	})
}

// GetMyTeacherProfile godoc
// @Summary Get my teacher profile
// @Description Get current user's teacher profile (Teacher users only)
// @Tags teacher-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with teacher data"
// @Failure 404 {object} map[string]string "Teacher profile not found"
// @Router /api/my-teacher-profile [get]
func (h *TeacherHandler) GetMyTeacherProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	teacher, err := h.teacherService.GetTeacherByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Teacher profile not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    teacher,
	})
}

// UpdateTeacher godoc
// @Summary Update teacher
// @Description Update teacher information
// @Tags teachers
// @Accept json
// @Produce json
// @Param id path int true "Teacher ID"
// @Param request body models.UpdateTeacherRequest true "Teacher update data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with updated teacher data"
// @Router /api/teachers/{id} [put]
func (h *TeacherHandler) UpdateTeacher(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid teacher ID",
		})
		return
	}

	var req models.UpdateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert to map for updates
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Salary != nil {
		updates["salary"] = *req.Salary
	}
	if req.Qualification != "" {
		updates["qualification"] = req.Qualification
	}
	if req.Experience != "" {
		updates["experience"] = req.Experience
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	updatedTeacher, err := h.teacherService.UpdateTeacher(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Teacher updated successfully",
		"data":    updatedTeacher,
	})
}

// UpdateMyTeacherProfile godoc
// @Summary Update my teacher profile
// @Description Update current user's teacher profile (Teacher users only)
// @Tags teacher-profile
// @Accept json
// @Produce json
// @Param request body models.UpdateTeacherRequest true "Teacher update data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with updated teacher data"
// @Router /api/my-teacher-profile [put]
func (h *TeacherHandler) UpdateMyTeacherProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Get teacher by user ID first
	teacher, err := h.teacherService.GetTeacherByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Teacher profile not found",
		})
		return
	}

	var req models.UpdateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert to map for updates
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Qualification != "" {
		updates["qualification"] = req.Qualification
	}
	if req.Experience != "" {
		updates["experience"] = req.Experience
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	updatedTeacher, err := h.teacherService.UpdateTeacher(teacher.ID, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Teacher profile updated successfully",
		"data":    updatedTeacher,
	})
}

// DeleteTeacher godoc
// @Summary Delete teacher
// @Description Delete a teacher (Admin only)
// @Tags teachers
// @Accept json
// @Produce json
// @Param id path int true "Teacher ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Router /api/teachers/{id} [delete]
func (h *TeacherHandler) DeleteTeacher(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid teacher ID",
		})
		return
	}

	err = h.teacherService.DeleteTeacher(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Teacher deleted successfully",
	})
}

// GetTeachersByBusiness godoc
// @Summary Get teachers by business
// @Description Get all teachers for a specific business
// @Tags teachers
// @Accept json
// @Produce json
// @Param businessId path int true "Business ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with teachers list"
// @Router /api/businesses/{businessId}/teachers [get]
func (h *TeacherHandler) GetTeachersByBusiness(c *gin.Context) {
	businessIDParam := c.Param("businessId")
	businessID, err := strconv.ParseUint(businessIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	var filters repository.TeacherFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid query parameters",
		})
		return
	}

	teachers, total, err := h.teacherService.GetTeachersByBusiness(uint(businessID), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get teachers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"teachers": teachers,
			"total":    total,
			"page":     filters.Page,
			"limit":    filters.Limit,
		},
	})
}

// ChangeTeacherStatus godoc
// @Summary Change teacher status
// @Description Change the status of a teacher
// @Tags teachers
// @Accept json
// @Produce json
// @Param id path int true "Teacher ID"
// @Param request body map[string]int true "Status data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Router /api/teachers/{id}/status [patch]
func (h *TeacherHandler) ChangeTeacherStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid teacher ID",
		})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required,min=0,max=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err = h.teacherService.ChangeTeacherStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Teacher status updated successfully",
	})
}

// SearchTeachers godoc
// @Summary Search teachers
// @Description Search teachers by name, qualification, or experience
// @Tags teachers
// @Accept json
// @Produce json
// @Param q query string true "Search term"
// @Param limit query int false "Maximum number of results" default(10)
// @Param business_id query int false "Filter by business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with search results"
// @Router /api/teachers/search [get]
func (h *TeacherHandler) SearchTeachers(c *gin.Context) {
	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Search term is required",
		})
		return
	}

	limitParam := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		limit = 10
	}

	var businessID uint
	businessIDParam := c.Query("business_id")
	if businessIDParam != "" {
		id, err := strconv.ParseUint(businessIDParam, 10, 32)
		if err == nil {
			businessID = uint(id)
		}
	}

	var teachers []models.TeacherResponse
	if businessID > 0 {
		teachers, err = h.teacherService.SearchTeachers(searchTerm, limit, businessID)
	} else {
		teachers, err = h.teacherService.SearchTeachers(searchTerm, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search teachers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"teachers":    teachers,
			"search_term": searchTerm,
			"total_found": len(teachers),
		},
	})
}

// GetTeacherStats godoc
// @Summary Get teacher statistics
// @Description Get teacher statistics
// @Tags teachers
// @Accept json
// @Produce json
// @Param business_id query int false "Filter by business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with statistics"
// @Router /api/teachers/stats [get]
func (h *TeacherHandler) GetTeacherStats(c *gin.Context) {
	var businessID uint
	businessIDParam := c.Query("business_id")
	if businessIDParam != "" {
		id, err := strconv.ParseUint(businessIDParam, 10, 32)
		if err == nil {
			businessID = uint(id)
		}
	}

	var stats map[string]interface{}
	var err error

	if businessID > 0 {
		stats, err = h.teacherService.GetTeacherStats(businessID)
	} else {
		stats, err = h.teacherService.GetTeacherStats()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get teacher statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// BulkUpdateTeacherStatus godoc
// @Summary Bulk update teacher status
// @Description Update status for multiple teachers
// @Tags teachers
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bulk update data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Router /api/teachers/bulk/status [post]
func (h *TeacherHandler) BulkUpdateTeacherStatus(c *gin.Context) {
	var req struct {
		TeacherIDs []uint `json:"teacher_ids" binding:"required"`
		Status     int    `json:"status" binding:"required,min=0,max=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err := h.teacherService.BulkUpdateTeacherStatus(req.TeacherIDs, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Teacher statuses updated successfully",
	})
}

// BulkUpdateSalary godoc
// @Summary Bulk update teacher salary
// @Description Update salary for multiple teachers
// @Tags teachers
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bulk salary update data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Router /api/teachers/bulk/salary [post]
func (h *TeacherHandler) BulkUpdateSalary(c *gin.Context) {
	var req struct {
		TeacherIDs []uint  `json:"teacher_ids" binding:"required"`
		Salary     float64 `json:"salary" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err := h.teacherService.BulkUpdateSalary(req.TeacherIDs, req.Salary)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Teacher salaries updated successfully",
	})
}

// GetActiveTeachers godoc
// @Summary Get active teachers
// @Description Get all active teachers
// @Tags teachers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with active teachers list"
// @Router /api/teachers/active [get]
func (h *TeacherHandler) GetActiveTeachers(c *gin.Context) {
	teachers, err := h.teacherService.GetActiveTeachers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get active teachers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    teachers,
	})
}

// GetInactiveTeachers godoc
// @Summary Get inactive teachers
// @Description Get all inactive teachers
// @Tags teachers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with inactive teachers list"
// @Router /api/teachers/inactive [get]
func (h *TeacherHandler) GetInactiveTeachers(c *gin.Context) {
	teachers, err := h.teacherService.GetInactiveTeachers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get inactive teachers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    teachers,
	})
}

// GetSalaryStats godoc
// @Summary Get salary statistics
// @Description Get teacher salary statistics
// @Tags teachers
// @Accept json
// @Produce json
// @Param business_id query int false "Filter by business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with salary statistics"
// @Router /api/teachers/stats/salary [get]
func (h *TeacherHandler) GetSalaryStats(c *gin.Context) {
	var businessID uint
	businessIDParam := c.Query("business_id")
	if businessIDParam != "" {
		id, err := strconv.ParseUint(businessIDParam, 10, 32)
		if err == nil {
			businessID = uint(id)
		}
	}

	var stats map[string]interface{}
	var err error

	if businessID > 0 {
		stats, err = h.teacherService.GetSalaryStats(businessID)
	} else {
		stats, err = h.teacherService.GetSalaryStats()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get salary statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetQualificationStats godoc
// @Summary Get qualification statistics
// @Description Get teacher qualification statistics
// @Tags teachers
// @Accept json
// @Produce json
// @Param business_id query int false "Filter by business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with qualification statistics"
// @Router /api/teachers/stats/qualifications [get]
func (h *TeacherHandler) GetQualificationStats(c *gin.Context) {
	var businessID uint
	businessIDParam := c.Query("business_id")
	if businessIDParam != "" {
		id, err := strconv.ParseUint(businessIDParam, 10, 32)
		if err == nil {
			businessID = uint(id)
		}
	}

	var stats map[string]int64
	var err error

	if businessID > 0 {
		stats, err = h.teacherService.GetQualificationStats(businessID)
	} else {
		stats, err = h.teacherService.GetQualificationStats()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get qualification statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
