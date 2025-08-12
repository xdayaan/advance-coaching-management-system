package repository

import (
	"backend/internal/models"
	"backend/pkg/database"
	"fmt"

	"gorm.io/gorm"
)

type PackageRepository interface {
	// Basic CRUD operations
	Create(pkg *models.Package) error
	GetByID(id uint) (*models.Package, error)
	GetByName(name string) (*models.Package, error)
	GetAll(filters PackageFilters) ([]models.Package, int64, error)
	Update(pkg *models.Package) error
	Delete(id uint) error

	// Status operations
	GetActivePackages() ([]models.Package, error)
	GetInactivePackages() ([]models.Package, error)
	UpdatePackageStatus(packageID uint, status int) error

	// Validation and utility
	PackageNameExists(name string, excludePackageID ...uint) (bool, error)
	GetPackagesCount() (int64, error)

	// Price and period operations
	GetPackagesByPriceRange(minPrice, maxPrice float64) ([]models.Package, error)
	GetPackagesByValidationPeriod(minDays, maxDays int) ([]models.Package, error)

	// Bulk operations
	BulkUpdateStatus(packageIDs []uint, status int) error
	BulkDelete(packageIDs []uint) error

	// Advanced queries
	SearchPackages(searchTerm string, limit int) ([]models.Package, error)
	GetRecentPackages(limit int) ([]models.Package, error)
	GetPackagesByDateRange(startDate, endDate string) ([]models.Package, error)

	// Statistics
	GetPackageStats() (map[string]interface{}, error)
	GetPriceStatistics() (map[string]float64, error)
}

type PackageFilters struct {
	Status    *int    `form:"status" json:"status"`
	MinPrice  float64 `form:"min_price" json:"min_price"`
	MaxPrice  float64 `form:"max_price" json:"max_price"`
	MinPeriod int     `form:"min_period" json:"min_period"`
	MaxPeriod int     `form:"max_period" json:"max_period"`
	Search    string  `form:"search" json:"search"`
	Page      int     `form:"page" json:"page"`
	Limit     int     `form:"limit" json:"limit"`
	SortBy    string  `form:"sort_by" json:"sort_by"`       // created_on, name, price, validation_period
	SortOrder string  `form:"sort_order" json:"sort_order"` // asc, desc
}

type packageRepository struct {
	db *gorm.DB
}

func NewPackageRepository() PackageRepository {
	return &packageRepository{
		db: database.DB,
	}
}

// Basic CRUD operations

func (r *packageRepository) Create(pkg *models.Package) error {
	if pkg == nil {
		return fmt.Errorf("package cannot be nil")
	}
	return r.db.Create(pkg).Error
}

func (r *packageRepository) GetByID(id uint) (*models.Package, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid package ID")
	}

	var pkg models.Package
	err := r.db.First(&pkg, id).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *packageRepository) GetByName(name string) (*models.Package, error) {
	if name == "" {
		return nil, fmt.Errorf("package name cannot be empty")
	}

	var pkg models.Package
	err := r.db.Where("name = ?", name).First(&pkg).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *packageRepository) GetAll(filters PackageFilters) ([]models.Package, int64, error) {
	var packages []models.Package
	var total int64

	query := r.db.Model(&models.Package{})

	// Apply filters
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.MinPrice > 0 {
		query = query.Where("price >= ?", filters.MinPrice)
	}

	if filters.MaxPrice > 0 {
		query = query.Where("price <= ?", filters.MaxPrice)
	}

	if filters.MinPeriod > 0 {
		query = query.Where("validation_period >= ?", filters.MinPeriod)
	}

	if filters.MaxPeriod > 0 {
		query = query.Where("validation_period <= ?", filters.MaxPeriod)
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total first (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_on DESC" // default
	if filters.SortBy != "" {
		validSortFields := map[string]bool{
			"created_on":        true,
			"updated_on":        true,
			"name":              true,
			"price":             true,
			"validation_period": true,
			"status":            true,
		}
		if validSortFields[filters.SortBy] {
			sortOrder := "DESC"
			if filters.SortOrder == "asc" {
				sortOrder = "ASC"
			}
			orderBy = fmt.Sprintf("%s %s", filters.SortBy, sortOrder)
		}
	}
	query = query.Order(orderBy)

	// Apply pagination
	if filters.Limit > 0 {
		offset := 0
		if filters.Page > 1 {
			offset = (filters.Page - 1) * filters.Limit
		}
		query = query.Offset(offset).Limit(filters.Limit)
	}

	err := query.Find(&packages).Error
	return packages, total, err
}

func (r *packageRepository) Update(pkg *models.Package) error {
	if pkg == nil {
		return fmt.Errorf("package cannot be nil")
	}
	if pkg.ID == 0 {
		return fmt.Errorf("package ID cannot be zero")
	}
	return r.db.Save(pkg).Error
}

func (r *packageRepository) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid package ID")
	}
	return r.db.Delete(&models.Package{}, id).Error
}

// Status operations

func (r *packageRepository) GetActivePackages() ([]models.Package, error) {
	var packages []models.Package
	err := r.db.Where("status = 1").Order("created_on DESC").Find(&packages).Error
	return packages, err
}

func (r *packageRepository) GetInactivePackages() ([]models.Package, error) {
	var packages []models.Package
	err := r.db.Where("status = 0").Order("created_on DESC").Find(&packages).Error
	return packages, err
}

func (r *packageRepository) UpdatePackageStatus(packageID uint, status int) error {
	if packageID == 0 {
		return fmt.Errorf("invalid package ID")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Package{}).Where("id = ?", packageID).Update("status", status).Error
}

// Validation and utility

func (r *packageRepository) PackageNameExists(name string, excludePackageID ...uint) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("package name cannot be empty")
	}

	var count int64
	query := r.db.Model(&models.Package{}).Where("name = ?", name)

	if len(excludePackageID) > 0 && excludePackageID[0] > 0 {
		query = query.Where("id != ?", excludePackageID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *packageRepository) GetPackagesCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Package{}).Count(&count).Error
	return count, err
}

// Price and period operations

func (r *packageRepository) GetPackagesByPriceRange(minPrice, maxPrice float64) ([]models.Package, error) {
	var packages []models.Package
	query := r.db.Where("status = 1")

	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}

	err := query.Order("price ASC").Find(&packages).Error
	return packages, err
}

func (r *packageRepository) GetPackagesByValidationPeriod(minDays, maxDays int) ([]models.Package, error) {
	var packages []models.Package
	query := r.db.Where("status = 1")

	if minDays > 0 {
		query = query.Where("validation_period >= ?", minDays)
	}
	if maxDays > 0 {
		query = query.Where("validation_period <= ?", maxDays)
	}

	err := query.Order("validation_period ASC").Find(&packages).Error
	return packages, err
}

// Bulk operations

func (r *packageRepository) BulkUpdateStatus(packageIDs []uint, status int) error {
	if len(packageIDs) == 0 {
		return fmt.Errorf("no package IDs provided")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Package{}).Where("id IN ?", packageIDs).Update("status", status).Error
}

func (r *packageRepository) BulkDelete(packageIDs []uint) error {
	if len(packageIDs) == 0 {
		return fmt.Errorf("no package IDs provided")
	}

	return r.db.Where("id IN ?", packageIDs).Delete(&models.Package{}).Error
}

// Advanced queries

func (r *packageRepository) SearchPackages(searchTerm string, limit int) ([]models.Package, error) {
	if searchTerm == "" {
		return []models.Package{}, nil
	}

	var packages []models.Package
	query := r.db.Where("name ILIKE ? OR description ILIKE ?", "%"+searchTerm+"%", "%"+searchTerm+"%").
		Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&packages).Error
	return packages, err
}

func (r *packageRepository) GetRecentPackages(limit int) ([]models.Package, error) {
	var packages []models.Package
	query := r.db.Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(10) // default limit
	}

	err := query.Find(&packages).Error
	return packages, err
}

func (r *packageRepository) GetPackagesByDateRange(startDate, endDate string) ([]models.Package, error) {
	if startDate == "" || endDate == "" {
		return nil, fmt.Errorf("start date and end date cannot be empty")
	}

	var packages []models.Package
	err := r.db.Where("created_on BETWEEN ? AND ?", startDate, endDate).
		Order("created_on DESC").Find(&packages).Error
	return packages, err
}

// Statistics

func (r *packageRepository) GetPackageStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total packages
	var totalPackages int64
	if err := r.db.Model(&models.Package{}).Count(&totalPackages).Error; err != nil {
		return nil, err
	}
	stats["total_packages"] = totalPackages

	// Active packages
	var activePackages int64
	if err := r.db.Model(&models.Package{}).Where("status = 1").Count(&activePackages).Error; err != nil {
		return nil, err
	}
	stats["active_packages"] = activePackages

	// Inactive packages
	var inactivePackages int64
	if err := r.db.Model(&models.Package{}).Where("status = 0").Count(&inactivePackages).Error; err != nil {
		return nil, err
	}
	stats["inactive_packages"] = inactivePackages

	return stats, nil
}

func (r *packageRepository) GetPriceStatistics() (map[string]float64, error) {
	stats := make(map[string]float64)

	var result struct {
		MinPrice float64
		MaxPrice float64
		AvgPrice float64
	}

	err := r.db.Model(&models.Package{}).Where("status = 1").
		Select("MIN(price) as min_price, MAX(price) as max_price, AVG(price) as avg_price").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	stats["min_price"] = result.MinPrice
	stats["max_price"] = result.MaxPrice
	stats["avg_price"] = result.AvgPrice

	return stats, nil
}
