package handlers

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BusinessHandler struct {
	businessService services.BusinessService
}

func NewBusinessHandler(businessService services.BusinessService) *BusinessHandler {
	return &BusinessHandler{
		businessService: businessService,
	}
}

// GetBusinessBySlug godoc
// @Summary Get business by slug (Public)
// @Description Get a specific business by slug (no authentication required)
// @Tags businesses
// @Accept json
// @Produce json
// @Param slug path string true "Business slug"
// @Success 200 {object} map[string]interface{} "Success response with business data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Business not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/business/{slug} [get]
func (h *BusinessHandler) GetBusinessBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Slug parameter is required",
		})
		return
	}

	business, err := h.businessService.GetBusinessBySlug(slug)
	if err != nil {
		if err.Error() == "business not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Business not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get business",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    business,
	})
}

// GetMyBusiness godoc
// @Summary Get my business profile
// @Description Get current user's business profile (Business users only)
// @Tags business-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with business data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Business profile not found"
// @Router /api/my-business [get]
func (h *BusinessHandler) GetMyBusiness(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	business, err := h.businessService.GetBusinessByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Business profile not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    business,
	})
}

// UpdateMyBusiness godoc
// @Summary Update my business profile
// @Description Update current user's business profile (Business users only)
// @Tags business-profile
// @Accept json
// @Produce json
// @Param request body models.UpdateBusinessRequest true "Business update data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with updated business data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Business profile not found"
// @Router /api/my-business [put]
func (h *BusinessHandler) UpdateMyBusiness(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Get business by user ID first
	business, err := h.businessService.GetBusinessByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Business profile not found",
		})
		return
	}

	var req models.UpdateBusinessRequest
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
	if req.Slug != "" {
		updates["slug"] = req.Slug
	}
	if req.OwnerName != "" {
		updates["owner_name"] = req.OwnerName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.PackageID != nil {
		updates["package_id"] = req.PackageID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	updatedBusiness, err := h.businessService.UpdateBusiness(business.ID, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Business profile updated successfully",
		"data":    updatedBusiness,
	})
}

// CreateBusiness godoc
// @Summary Create a new business
// @Description Create a new business with the provided information (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param request body models.CreateBusinessRequest true "Business data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Success response with business data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/businesses [post]
func (h *BusinessHandler) CreateBusiness(c *gin.Context) {
	var req models.CreateBusinessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	business, err := h.businessService.CreateBusiness(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Business created successfully",
		"data":    business,
	})
}

// GetBusinesses godoc
// @Summary Get all businesses
// @Description Get all businesses with pagination and filters (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query int false "Filter by status (0=inactive, 1=active)"
// @Param package_id query int false "Filter by package ID"
// @Param location query string false "Filter by location"
// @Param search query string false "Search in name, owner name, email, location, or slug"
// @Param sort_by query string false "Sort by field (name, owner_name, email, location, status, created_on)"
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with businesses list"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses [get]
func (h *BusinessHandler) GetBusinesses(c *gin.Context) {
	var filters repository.BusinessFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	businesses, total, err := h.businessService.GetBusinesses(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get businesses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"businesses": businesses,
			"total":      total,
			"page":       filters.Page,
			"limit":      filters.Limit,
		},
	})
}

// GetBusiness godoc
// @Summary Get business by ID
// @Description Get a specific business by ID (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with business data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Business not found"
// @Router /api/businesses/{id} [get]
func (h *BusinessHandler) GetBusiness(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	business, err := h.businessService.GetBusinessByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Business not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    business,
	})
}

// UpdateBusiness godoc
// @Summary Update business
// @Description Update business information (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param request body models.UpdateBusinessRequest true "Business update data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with updated business data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Business not found"
// @Router /api/businesses/{id} [put]
func (h *BusinessHandler) UpdateBusiness(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	var req models.UpdateBusinessRequest
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
	if req.Slug != "" {
		updates["slug"] = req.Slug
	}
	if req.OwnerName != "" {
		updates["owner_name"] = req.OwnerName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.PackageID != nil {
		updates["package_id"] = req.PackageID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	updatedBusiness, err := h.businessService.UpdateBusiness(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Business updated successfully",
		"data":    updatedBusiness,
	})
}

// DeleteBusiness godoc
// @Summary Delete business
// @Description Delete a business (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Business not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/{id} [delete]
func (h *BusinessHandler) DeleteBusiness(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	err = h.businessService.DeleteBusiness(uint(id))
	if err != nil {
		if err.Error() == "business not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Business not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete business",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Business deleted successfully",
	})
}

// ChangeBusinessStatus godoc
// @Summary Change business status
// @Description Change the status of a business (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param request body map[string]int true "Status data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Business not found"
// @Router /api/businesses/{id}/status [patch]
func (h *BusinessHandler) ChangeBusinessStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
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

	err = h.businessService.ChangeBusinessStatus(uint(id), req.Status)
	if err != nil {
		if err.Error() == "business not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Business not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Business status updated successfully",
	})
}

// AssignPackage godoc
// @Summary Assign package to business
// @Description Assign a package to a business (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param request body models.AssignPackageRequest true "Package assignment data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/businesses/{id}/assign-package [post]
func (h *BusinessHandler) AssignPackage(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	var req models.AssignPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err = h.businessService.AssignPackage(uint(id), req.PackageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Package assigned successfully",
	})
}

// RemovePackage godoc
// @Summary Remove package from business
// @Description Remove package assignment from a business (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/businesses/{id}/remove-package [delete]
func (h *BusinessHandler) RemovePackage(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid business ID",
		})
		return
	}

	err = h.businessService.RemovePackage(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Package removed successfully",
	})
}

// SearchBusinesses godoc
// @Summary Search businesses
// @Description Search businesses by name, owner name, email, location, or slug (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param q query string true "Search term"
// @Param limit query int false "Maximum number of results" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with search results"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/search [get]
func (h *BusinessHandler) SearchBusinesses(c *gin.Context) {
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

	businesses, err := h.businessService.SearchBusinesses(searchTerm, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search businesses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"businesses":  businesses,
			"search_term": searchTerm,
			"total_found": len(businesses),
		},
	})
}

// GetBusinessStats godoc
// @Summary Get business statistics
// @Description Get business statistics (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/stats [get]
func (h *BusinessHandler) GetBusinessStats(c *gin.Context) {
	stats, err := h.businessService.GetBusinessStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get business statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetActiveBusinesses godoc
// @Summary Get active businesses
// @Description Get all active businesses (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with active businesses list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/active [get]
func (h *BusinessHandler) GetActiveBusinesses(c *gin.Context) {
	businesses, err := h.businessService.GetActiveBusinesses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get active businesses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    businesses,
	})
}

// GetInactiveBusinesses godoc
// @Summary Get inactive businesses
// @Description Get all inactive businesses (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with inactive businesses list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/inactive [get]
func (h *BusinessHandler) GetInactiveBusinesses(c *gin.Context) {
	businesses, err := h.businessService.GetInactiveBusinesses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get inactive businesses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    businesses,
	})
}

// GetBusinessesByPackage godoc
// @Summary Get businesses by package
// @Description Get all businesses assigned to a specific package (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param packageId path int true "Package ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with businesses list"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/package/{packageId} [get]
func (h *BusinessHandler) GetBusinessesByPackage(c *gin.Context) {
	packageIDParam := c.Param("packageId")
	packageID, err := strconv.ParseUint(packageIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid package ID",
		})
		return
	}

	businesses, err := h.businessService.GetBusinessesByPackage(uint(packageID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get businesses by package",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    businesses,
	})
}

// GetBusinessesWithoutPackage godoc
// @Summary Get businesses without package
// @Description Get all businesses that don't have any package assigned (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with businesses list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/no-package [get]
func (h *BusinessHandler) GetBusinessesWithoutPackage(c *gin.Context) {
	businesses, err := h.businessService.GetBusinessesWithoutPackage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get businesses without package",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    businesses,
	})
}

// GetLocationStats godoc
// @Summary Get location statistics
// @Description Get business location statistics (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with location statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/stats/locations [get]
func (h *BusinessHandler) GetLocationStats(c *gin.Context) {
	stats, err := h.businessService.GetLocationStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get location statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetPackageDistribution godoc
// @Summary Get package distribution statistics
// @Description Get statistics of package distribution among businesses (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with package distribution statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/stats/packages [get]
func (h *BusinessHandler) GetPackageDistribution(c *gin.Context) {
	stats, err := h.businessService.GetPackageDistribution()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get package distribution",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// BulkUpdateStatus godoc
// @Summary Bulk update business status
// @Description Update status for multiple businesses (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bulk update data with business_ids and status"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/businesses/bulk/status [post]
func (h *BusinessHandler) BulkUpdateStatus(c *gin.Context) {
	var req struct {
		BusinessIDs []uint `json:"business_ids" binding:"required"`
		Status      int    `json:"status" binding:"required,min=0,max=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err := h.businessService.BulkUpdateBusinessStatus(req.BusinessIDs, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Business statuses updated successfully",
	})
}

// BulkAssignPackage godoc
// @Summary Bulk assign package to businesses
// @Description Assign a package to multiple businesses (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bulk assignment data with business_ids and package_id"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/businesses/bulk/assign-package [post]
func (h *BusinessHandler) BulkAssignPackage(c *gin.Context) {
	var req struct {
		BusinessIDs []uint `json:"business_ids" binding:"required"`
		PackageID   uint   `json:"package_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err := h.businessService.BulkAssignPackage(req.BusinessIDs, req.PackageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Package assigned to businesses successfully",
	})
}

// GetBusinessesByLocation godoc
// @Summary Get businesses by location
// @Description Get all businesses filtered by location (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Param location query string true "Location filter"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with businesses list"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/by-location [get]
func (h *BusinessHandler) GetBusinessesByLocation(c *gin.Context) {
	location := c.Query("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Location parameter is required",
		})
		return
	}

	businesses, err := h.businessService.GetBusinessesByLocation(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get businesses by location",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    businesses,
	})
}

// GetBusinessLocations godoc
// @Summary Get all business locations
// @Description Get a list of all unique business locations (Admin only)
// @Tags businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with locations list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/businesses/locations [get]
func (h *BusinessHandler) GetBusinessLocations(c *gin.Context) {
	locations, err := h.businessService.GetBusinessLocations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get business locations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    locations,
	})
}
