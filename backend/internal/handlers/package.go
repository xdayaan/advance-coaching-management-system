package handlers

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PackageHandler struct {
	packageService services.PackageService
}

func NewPackageHandler(packageService services.PackageService) *PackageHandler {
	return &PackageHandler{
		packageService: packageService,
	}
}

// CreatePackage godoc
// @Summary Create a new package
// @Description Create a new package with the provided information (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Param request body models.CreatePackageRequest true "Package data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Success response with package data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 409 {object} map[string]string "Package name already exists"
// @Router /api/packages [post]
func (h *PackageHandler) CreatePackage(c *gin.Context) {
	var req models.CreatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	pkg, err := h.packageService.CreatePackage(req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "name already exists") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "cannot be negative") ||
			strings.Contains(err.Error(), "must be greater than") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Package created successfully",
		"data":    pkg,
	})
}

// GetPackages godoc
// @Summary Get all packages
// @Description Get all packages with pagination and filters
// @Tags packages
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query int false "Filter by status (0=inactive, 1=active)"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param min_period query int false "Minimum validation period filter (days)"
// @Param max_period query int false "Maximum validation period filter (days)"
// @Param search query string false "Search in name or description"
// @Param sort_by query string false "Sort by field (name, price, validation_period, created_on)"
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with packages list"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/packages [get]
func (h *PackageHandler) GetPackages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Handle status filter
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		if statusVal, err := strconv.Atoi(statusStr); err == nil {
			if statusVal == 0 || statusVal == 1 {
				status = &statusVal
			}
		}
	}

	// Handle price filters
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)

	// Handle period filters
	minPeriod, _ := strconv.Atoi(c.Query("min_period"))
	maxPeriod, _ := strconv.Atoi(c.Query("max_period"))

	filters := repository.PackageFilters{
		Status:    status,
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
		MinPeriod: minPeriod,
		MaxPeriod: maxPeriod,
		Search:    c.Query("search"),
		Page:      page,
		Limit:     limit,
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	packages, total, err := h.packageService.GetPackages(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate pagination info
	totalPages := (int(total) + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"data": packages,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

// GetPackage godoc
// @Summary Get package by ID
// @Description Get a specific package by ID
// @Tags packages
// @Accept json
// @Produce json
// @Param id path int true "Package ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with package data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Package not found"
// @Router /api/packages/{id} [get]
func (h *PackageHandler) GetPackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}

	pkg, err := h.packageService.GetPackageByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pkg})
}

// UpdatePackage godoc
// @Summary Update package
// @Description Update package information (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Param id path int true "Package ID"
// @Param request body map[string]interface{} true "Update data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with updated package data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Package not found"
// @Failure 409 {object} map[string]string "Package name already exists"
// @Router /api/packages/{id} [put]
func (h *PackageHandler) UpdatePackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	pkg, err := h.packageService.UpdatePackage(uint(id), updates)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "name already exists") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "cannot be negative") ||
			strings.Contains(err.Error(), "must be greater than") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Package updated successfully",
		"data":    pkg,
	})
}

// DeletePackage godoc
// @Summary Delete package
// @Description Delete a package (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Param id path int true "Package ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Package not found"
// @Router /api/packages/{id} [delete]
func (h *PackageHandler) DeletePackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}

	if err := h.packageService.DeletePackage(uint(id)); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Package deleted successfully"})
}

// GetActivePackages godoc
// @Summary Get active packages
// @Description Get all active packages
// @Tags packages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with active packages list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/packages/active [get]
func (h *PackageHandler) GetActivePackages(c *gin.Context) {
	packages, err := h.packageService.GetActivePackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  packages,
		"count": len(packages),
	})
}

// GetInactivePackages godoc
// @Summary Get inactive packages
// @Description Get all inactive packages (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with inactive packages list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/packages/inactive [get]
func (h *PackageHandler) GetInactivePackages(c *gin.Context) {
	packages, err := h.packageService.GetInactivePackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  packages,
		"count": len(packages),
	})
}

// ChangePackageStatus godoc
// @Summary Change package status
// @Description Change the status of a package (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Param id path int true "Package ID"
// @Param request body map[string]int true "Status data"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Package not found"
// @Router /api/packages/{id}/status [patch]
func (h *PackageHandler) ChangePackageStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.packageService.ChangePackageStatus(uint(id), req.Status); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	statusText := "inactive"
	if req.Status == 1 {
		statusText = "active"
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Package status changed to %s successfully", statusText)})
}

// GetPackageStats godoc
// @Summary Get package statistics
// @Description Get package statistics (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/packages/stats [get]
func (h *PackageHandler) GetPackageStats(c *gin.Context) {
	stats, err := h.packageService.GetPackageStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetPriceStatistics godoc
// @Summary Get price statistics
// @Description Get package price statistics (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with price statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/packages/stats/prices [get]
func (h *PackageHandler) GetPriceStatistics(c *gin.Context) {
	stats, err := h.packageService.GetPriceStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetPackagesByPriceRange godoc
// @Summary Get packages by price range
// @Description Get packages within a specific price range
// @Tags packages
// @Accept json
// @Produce json
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with packages list"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/packages/price-range [get]
func (h *PackageHandler) GetPackagesByPriceRange(c *gin.Context) {
	minPriceStr := c.Query("min_price")
	maxPriceStr := c.Query("max_price")

	var minPrice, maxPrice float64
	var err error

	if minPriceStr != "" {
		minPrice, err = strconv.ParseFloat(minPriceStr, 64)
		if err != nil || minPrice < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid minimum price"})
			return
		}
	}

	if maxPriceStr != "" {
		maxPrice, err = strconv.ParseFloat(maxPriceStr, 64)
		if err != nil || maxPrice < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maximum price"})
			return
		}
	}

	if minPrice > 0 && maxPrice > 0 && minPrice > maxPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum price cannot be greater than maximum price"})
		return
	}

	packages, err := h.packageService.GetPackagesByPriceRange(minPrice, maxPrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      packages,
		"count":     len(packages),
		"min_price": minPrice,
		"max_price": maxPrice,
	})
}

// BulkUpdatePackageStatus godoc
// @Summary Bulk update package status
// @Description Update status for multiple packages (Admin/Business only)
// @Tags packages
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bulk update data with package_ids and status"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/packages/bulk/status [patch]
func (h *PackageHandler) BulkUpdatePackageStatus(c *gin.Context) {
	var req struct {
		PackageIDs []uint `json:"package_ids" binding:"required"`
		Status     int    `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if len(req.PackageIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Package IDs are required"})
		return
	}

	if req.Status < 0 || req.Status > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value. Must be 0 (inactive) or 1 (active)"})
		return
	}

	if err := h.packageService.BulkUpdatePackageStatus(req.PackageIDs, req.Status); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	statusText := "inactive"
	if req.Status == 1 {
		statusText = "active"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%d packages status changed to %s successfully", len(req.PackageIDs), statusText),
	})
}

// SearchPackages godoc
// @Summary Search packages
// @Description Search packages by name or description
// @Tags packages
// @Accept json
// @Produce json
// @Param q query string true "Search term"
// @Param limit query int false "Maximum number of results" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response with search results"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/packages/search [get]
func (h *PackageHandler) SearchPackages(c *gin.Context) {
	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	packages, err := h.packageService.SearchPackages(searchTerm, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        packages,
		"count":       len(packages),
		"search_term": searchTerm,
		"limit":       limit,
	})
}
