package handlers

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StudentHandler struct {
	studentService services.StudentService
}

func NewStudentHandler(studentService services.StudentService) *StudentHandler {
	return &StudentHandler{
		studentService: studentService,
	}
}

// CreateStudent godoc
// @Summary Create a new student
// @Description Create a new student (Admin/Business only)
// @Tags students
// @Accept json
// @Produce json
// @Param request body models.CreateStudentRequest true "Student data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Success response with student data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/students [post]
func (h *StudentHandler) CreateStudent(c *gin.Context) {
	var req models.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	student, err := h.studentService.CreateStudent(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Student created successfully",
		"data":    student,
	})
}

// GetStudents godoc
// @Summary Get all students
// @Description Get all students with pagination and filters (Admin only)
// @Tags students
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query int false "Filter by status (0=inactive, 1=active)"
// @Param business_id query int false "Filter by business ID"
// @Param guardian_name query string false "Filter by guardian name"
// @Param guardian_email query string false "Filter by guardian email"
// @Param search query string false "Search in name, guardian info"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with students list"
// @Router /api/students [get]
func (h *StudentHandler) GetStudents(c *gin.Context) {
	var filters repository.StudentFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	students, total, err := h.studentService.GetStudents(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"students": students,
			"total":    total,
			"page":     filters.Page,
			"limit":    filters.Limit,
		},
	})
}

// GetStudent godoc
// @Summary Get student by ID
// @Description Get a specific student by ID
// @Tags students
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with student data"
// @Failure 404 {object} map[string]string "Student not found"
// @Router /api/students/{id} [get]
func (h *StudentHandler) GetStudent(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid student ID",
		})
		return
	}

	student, err := h.studentService.GetStudentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Student not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    student,
	})
}

// GetMyStudentProfile godoc
// @Summary Get my student profile
// @Description Get current user's student profile (Student users only)
// @Tags student-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with student data"
// @Failure 404 {object} map[string]string "Student profile not found"
// @Router /api/my-student-profile [get]
func (h *StudentHandler) GetMyStudentProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	student, err := h.studentService.GetStudentByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Student profile not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    student,
	})
}

// UpdateStudent godoc
// @Summary Update student
// @Description Update student information
// @Tags students
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Param request body models.UpdateStudentRequest true "Student update data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with updated student data"
// @Router /api/students/{id} [put]
func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid student ID",
		})
		return
	}

	var req models.UpdateStudentRequest
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
	if req.GuardianName != "" {
		updates["guardian_name"] = req.GuardianName
	}
	if req.GuardianNumber != "" {
		updates["guardian_number"] = req.GuardianNumber
	}
	if req.GuardianEmail != "" {
		updates["guardian_email"] = req.GuardianEmail
	}
	if req.Information != nil {
		updates["information"] = req.Information
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	updatedStudent, err := h.studentService.UpdateStudent(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Student updated successfully",
		"data":    updatedStudent,
	})
}

// UpdateMyStudentProfile godoc
// @Summary Update my student profile
// @Description Update current user's student profile (Student users only)
// @Tags student-profile
// @Accept json
// @Produce json
// @Param request body models.UpdateStudentRequest true "Student update data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with updated student data"
// @Router /api/my-student-profile [put]
func (h *StudentHandler) UpdateMyStudentProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Get student by user ID first
	student, err := h.studentService.GetStudentByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Student profile not found",
		})
		return
	}

	var req models.UpdateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert to map for updates (students can update limited fields)
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.GuardianName != "" {
		updates["guardian_name"] = req.GuardianName
	}
	if req.GuardianNumber != "" {
		updates["guardian_number"] = req.GuardianNumber
	}
	if req.GuardianEmail != "" {
		updates["guardian_email"] = req.GuardianEmail
	}
	if req.Information != nil {
		updates["information"] = req.Information
	}

	updatedStudent, err := h.studentService.UpdateStudent(student.ID, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Student profile updated successfully",
		"data":    updatedStudent,
	})
}

// DeleteStudent godoc
// @Summary Delete student
// @Description Delete a student (Admin only)
// @Tags students
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Router /api/students/{id} [delete]
func (h *StudentHandler) DeleteStudent(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid student ID",
		})
		return
	}

	err = h.studentService.DeleteStudent(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Student deleted successfully",
	})
}

// GetStudentsByBusiness godoc
// @Summary Get students by business
// @Description Get all students for a specific business
// @Tags students
// @Accept json
// @Produce json
// @Param businessId path int true "Business ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with students list"
// @Router /api/businesses/{businessId}/students [get]
func (h *StudentHandler) GetStudentsByBusiness(c *gin.Context) {
	businessIDParam := c.Param("businessId")
	businessID, err := strconv.ParseUint(businessIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	var filters repository.StudentFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid query parameters",
		})
		return
	}

	students, total, err := h.studentService.GetStudentsByBusiness(uint(businessID), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"students": students,
			"total":    total,
			"page":     filters.Page,
			"limit":    filters.Limit,
		},
	})
}

// ChangeStudentStatus godoc
// @Summary Change student status
// @Description Change the status of a student
// @Tags students
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Param request body map[string]int true "Status data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Router /api/students/{id}/status [patch]
func (h *StudentHandler) ChangeStudentStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid student ID",
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

	err = h.studentService.ChangeStudentStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Student status updated successfully",
	})
}

// SearchStudents godoc
// @Summary Search students
// @Description Search students by name, guardian name, email, or number
// @Tags students
// @Accept json
// @Produce json
// @Param q query string true "Search term"
// @Param limit query int false "Maximum number of results" default(10)
// @Param business_id query int false "Filter by business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with search results"
// @Router /api/students/search [get]
func (h *StudentHandler) SearchStudents(c *gin.Context) {
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

	var students []models.StudentResponse
	if businessID > 0 {
		students, err = h.studentService.SearchStudents(searchTerm, limit, businessID)
	} else {
		students, err = h.studentService.SearchStudents(searchTerm, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"students":    students,
			"search_term": searchTerm,
			"total_found": len(students),
		},
	})
}

// GetStudentStats godoc
// @Summary Get student statistics
// @Description Get student statistics
// @Tags students
// @Accept json
// @Produce json
// @Param business_id query int false "Filter by business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with statistics"
// @Router /api/students/stats [get]
func (h *StudentHandler) GetStudentStats(c *gin.Context) {
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
		stats, err = h.studentService.GetStudentStats(businessID)
	} else {
		stats, err = h.studentService.GetStudentStats()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get student statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// BulkUpdateStudentStatus godoc
// @Summary Bulk update student status
// @Description Update status for multiple students
// @Tags students
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bulk update data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Router /api/students/bulk/status [post]
func (h *StudentHandler) BulkUpdateStudentStatus(c *gin.Context) {
	var req struct {
		StudentIDs []uint `json:"student_ids" binding:"required"`
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

	err := h.studentService.BulkUpdateStudentStatus(req.StudentIDs, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Student statuses updated successfully",
	})
}

// GetActiveStudents godoc
// @Summary Get active students
// @Description Get all active students
// @Tags students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with active students list"
// @Router /api/students/active [get]
func (h *StudentHandler) GetActiveStudents(c *gin.Context) {
	students, err := h.studentService.GetActiveStudents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get active students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    students,
	})
}

// GetInactiveStudents godoc
// @Summary Get inactive students
// @Description Get all inactive students
// @Tags students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with inactive students list"
// @Router /api/students/inactive [get]
func (h *StudentHandler) GetInactiveStudents(c *gin.Context) {
	students, err := h.studentService.GetInactiveStudents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get inactive students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    students,
	})
}

// GetGuardianStats godoc
// @Summary Get guardian statistics
// @Description Get student guardian statistics
// @Tags students
// @Accept json
// @Produce json
// @Param business_id query int false "Filter by business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with guardian statistics"
// @Router /api/students/stats/guardians [get]
func (h *StudentHandler) GetGuardianStats(c *gin.Context) {
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
		stats, err = h.studentService.GetGuardianStats(businessID)
	} else {
		stats, err = h.studentService.GetGuardianStats()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get guardian statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetActiveStudentsByBusiness godoc
// @Summary Get active students by business
// @Description Get all active students for a specific business
// @Tags students
// @Accept json
// @Produce json
// @Param businessId path int true "Business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with active students list"
// @Router /api/businesses/{businessId}/students/active [get]
func (h *StudentHandler) GetActiveStudentsByBusiness(c *gin.Context) {
	businessIDParam := c.Param("businessId")
	businessID, err := strconv.ParseUint(businessIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	students, err := h.studentService.GetActiveStudentsByBusiness(uint(businessID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get active students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    students,
	})
}

// GetInactiveStudentsByBusiness godoc
// @Summary Get inactive students by business
// @Description Get all inactive students for a specific business
// @Tags students
// @Accept json
// @Produce json
// @Param businessId path int true "Business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with inactive students list"
// @Router /api/businesses/{businessId}/students/inactive [get]
func (h *StudentHandler) GetInactiveStudentsByBusiness(c *gin.Context) {
	businessIDParam := c.Param("businessId")
	businessID, err := strconv.ParseUint(businessIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	students, err := h.studentService.GetInactiveStudentsByBusiness(uint(businessID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get inactive students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    students,
	})
}
