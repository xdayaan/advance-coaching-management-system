package repository

import (
	"backend/internal/models"
	"backend/pkg/database"
	"fmt"

	"gorm.io/gorm"
)

type BusinessRepository interface {
	// Basic CRUD operations
	Create(business *models.Business) error
	CreateWithTransaction(tx *gorm.DB, business *models.Business) error
	GetByID(id uint) (*models.Business, error)
	GetBySlug(slug string) (*models.Business, error)
	GetByUserID(userID uint) (*models.Business, error)
	GetByEmail(email string) (*models.Business, error)
	GetAll(filters BusinessFilters) ([]models.Business, int64, error)
	GetAllWithRelations(filters BusinessFilters) ([]models.Business, int64, error)
	Update(business *models.Business) error
	UpdateWithTransaction(tx *gorm.DB, business *models.Business) error
	Delete(id uint) error

	// Status operations
	UpdateBusinessStatus(businessID uint, status int) error
	GetActiveBusinesses() ([]models.Business, error)
	GetInactiveBusinesses() ([]models.Business, error)

	// Package operations
	AssignPackage(businessID, packageID uint) error
	RemovePackage(businessID uint) error
	GetBusinessesByPackage(packageID uint) ([]models.Business, error)
	GetBusinessesWithoutPackage() ([]models.Business, error)

	// Validation and utility
	BusinessEmailExists(email string, excludeBusinessID ...uint) (bool, error)
	BusinessSlugExists(slug string, excludeBusinessID ...uint) (bool, error)
	BusinessNameExists(name string, excludeBusinessID ...uint) (bool, error)

	// Search
	SearchBusinesses(searchTerm string, limit int) ([]models.Business, error)

	// Statistics
	GetBusinessStats() (map[string]interface{}, error)
	GetLocationStats() (map[string]int64, error)
	GetPackageDistribution() (map[string]int64, error)

	// Relationships
	GetBusinessWithRelations(id uint) (*models.Business, error)
	GetBySlugWithRelations(slug string) (*models.Business, error)

	// Bulk operations
	BulkUpdateStatus(businessIDs []uint, status int) error
	BulkAssignPackage(businessIDs []uint, packageID uint) error

	// Location operations
	GetBusinessesByLocation(location string) ([]models.Business, error)
	GetBusinessLocations() ([]string, error)

	// Transaction support
	BeginTransaction() *gorm.DB
}

type BusinessFilters struct {
	PackageID *uint  `form:"package_id" json:"package_id"`
	Status    *int   `form:"status" json:"status"`
	Location  string `form:"location" json:"location"`
	Search    string `form:"search" json:"search"`
	Page      int    `form:"page" json:"page"`
	Limit     int    `form:"limit" json:"limit"`
	SortBy    string `form:"sort_by" json:"sort_by"`
	SortOrder string `form:"sort_order" json:"sort_order"`
}

type businessRepository struct {
	db *gorm.DB
}

func NewBusinessRepository() BusinessRepository {
	return &businessRepository{
		db: database.DB,
	}
}

// Basic CRUD operations

func (r *businessRepository) Create(business *models.Business) error {
	if business == nil {
		return fmt.Errorf("business cannot be nil")
	}
	return r.db.Create(business).Error
}

func (r *businessRepository) CreateWithTransaction(tx *gorm.DB, business *models.Business) error {
	if business == nil {
		return fmt.Errorf("business cannot be nil")
	}
	return tx.Create(business).Error
}

func (r *businessRepository) GetByID(id uint) (*models.Business, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid business ID")
	}

	var business models.Business
	err := r.db.First(&business, id).Error
	if err != nil {
		return nil, err
	}
	return &business, nil
}

func (r *businessRepository) GetBySlug(slug string) (*models.Business, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	var business models.Business
	err := r.db.Where("slug = ?", slug).First(&business).Error
	if err != nil {
		return nil, err
	}
	return &business, nil
}

func (r *businessRepository) GetByUserID(userID uint) (*models.Business, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	var business models.Business
	err := r.db.Where("user_id = ?", userID).First(&business).Error
	if err != nil {
		return nil, err
	}
	return &business, nil
}

func (r *businessRepository) GetByEmail(email string) (*models.Business, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var business models.Business
	err := r.db.Where("email = ?", email).First(&business).Error
	if err != nil {
		return nil, err
	}
	return &business, nil
}

func (r *businessRepository) GetAll(filters BusinessFilters) ([]models.Business, int64, error) {
	var businesses []models.Business
	var total int64

	query := r.db.Model(&models.Business{})

	// Apply filters
	if filters.PackageID != nil {
		if *filters.PackageID == 0 {
			query = query.Where("package_id IS NULL")
		} else {
			query = query.Where("package_id = ?", *filters.PackageID)
		}
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.Location != "" {
		query = query.Where("location ILIKE ?", "%"+filters.Location+"%")
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR owner_name ILIKE ? OR email ILIKE ? OR location ILIKE ? OR slug ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total first (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_on DESC" // default
	if filters.SortBy != "" {
		validSortFields := map[string]bool{
			"created_on": true,
			"updated_on": true,
			"name":       true,
			"owner_name": true,
			"email":      true,
			"location":   true,
			"status":     true,
			"slug":       true,
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

	err := query.Find(&businesses).Error
	return businesses, total, err
}

func (r *businessRepository) GetAllWithRelations(filters BusinessFilters) ([]models.Business, int64, error) {
	var businesses []models.Business
	var total int64

	query := r.db.Model(&models.Business{}).Preload("User").Preload("Package")

	// Apply filters
	if filters.PackageID != nil {
		if *filters.PackageID == 0 {
			query = query.Where("package_id IS NULL")
		} else {
			query = query.Where("package_id = ?", *filters.PackageID)
		}
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.Location != "" {
		query = query.Where("location ILIKE ?", "%"+filters.Location+"%")
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR owner_name ILIKE ? OR email ILIKE ? OR location ILIKE ? OR slug ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total first (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_on DESC"
	if filters.SortBy != "" {
		validSortFields := map[string]bool{
			"created_on": true,
			"updated_on": true,
			"name":       true,
			"owner_name": true,
			"email":      true,
			"location":   true,
			"status":     true,
			"slug":       true,
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

	err := query.Find(&businesses).Error
	return businesses, total, err
}

func (r *businessRepository) Update(business *models.Business) error {
	if business == nil {
		return fmt.Errorf("business cannot be nil")
	}
	if business.ID == 0 {
		return fmt.Errorf("business ID cannot be zero")
	}
	return r.db.Save(business).Error
}

func (r *businessRepository) UpdateWithTransaction(tx *gorm.DB, business *models.Business) error {
	if business == nil {
		return fmt.Errorf("business cannot be nil")
	}
	if business.ID == 0 {
		return fmt.Errorf("business ID cannot be zero")
	}
	return tx.Save(business).Error
}

func (r *businessRepository) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid business ID")
	}
	return r.db.Delete(&models.Business{}, id).Error
}

// Status operations

func (r *businessRepository) UpdateBusinessStatus(businessID uint, status int) error {
	if businessID == 0 {
		return fmt.Errorf("invalid business ID")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Business{}).Where("id = ?", businessID).Update("status", status).Error
}

func (r *businessRepository) GetActiveBusinesses() ([]models.Business, error) {
	var businesses []models.Business
	err := r.db.Where("status = 1").Find(&businesses).Error
	return businesses, err
}

func (r *businessRepository) GetInactiveBusinesses() ([]models.Business, error) {
	var businesses []models.Business
	err := r.db.Where("status = 0").Find(&businesses).Error
	return businesses, err
}

// Package operations

func (r *businessRepository) AssignPackage(businessID, packageID uint) error {
	if businessID == 0 || packageID == 0 {
		return fmt.Errorf("invalid business ID or package ID")
	}

	return r.db.Model(&models.Business{}).Where("id = ?", businessID).Update("package_id", packageID).Error
}

func (r *businessRepository) RemovePackage(businessID uint) error {
	if businessID == 0 {
		return fmt.Errorf("invalid business ID")
	}

	return r.db.Model(&models.Business{}).Where("id = ?", businessID).Update("package_id", nil).Error
}

func (r *businessRepository) GetBusinessesByPackage(packageID uint) ([]models.Business, error) {
	if packageID == 0 {
		return nil, fmt.Errorf("invalid package ID")
	}

	var businesses []models.Business
	err := r.db.Where("package_id = ?", packageID).Find(&businesses).Error
	return businesses, err
}

func (r *businessRepository) GetBusinessesWithoutPackage() ([]models.Business, error) {
	var businesses []models.Business
	err := r.db.Where("package_id IS NULL").Find(&businesses).Error
	return businesses, err
}

// Validation and utility

func (r *businessRepository) BusinessEmailExists(email string, excludeBusinessID ...uint) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("email cannot be empty")
	}

	var count int64
	query := r.db.Model(&models.Business{}).Where("email = ?", email)

	if len(excludeBusinessID) > 0 && excludeBusinessID[0] > 0 {
		query = query.Where("id != ?", excludeBusinessID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *businessRepository) BusinessSlugExists(slug string, excludeBusinessID ...uint) (bool, error) {
	if slug == "" {
		return false, fmt.Errorf("slug cannot be empty")
	}

	var count int64
	query := r.db.Model(&models.Business{}).Where("slug = ?", slug)

	if len(excludeBusinessID) > 0 && excludeBusinessID[0] > 0 {
		query = query.Where("id != ?", excludeBusinessID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *businessRepository) BusinessNameExists(name string, excludeBusinessID ...uint) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name cannot be empty")
	}

	var count int64
	query := r.db.Model(&models.Business{}).Where("name = ?", name)

	if len(excludeBusinessID) > 0 && excludeBusinessID[0] > 0 {
		query = query.Where("id != ?", excludeBusinessID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// Search

func (r *businessRepository) SearchBusinesses(searchTerm string, limit int) ([]models.Business, error) {
	if searchTerm == "" {
		return []models.Business{}, nil
	}

	var businesses []models.Business
	query := r.db.Where("name ILIKE ? OR owner_name ILIKE ? OR email ILIKE ? OR location ILIKE ? OR slug ILIKE ?",
		"%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%").
		Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&businesses).Error
	return businesses, err
}

// Statistics

func (r *businessRepository) GetBusinessStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total businesses
	var totalBusinesses int64
	if err := r.db.Model(&models.Business{}).Count(&totalBusinesses).Error; err != nil {
		return nil, err
	}
	stats["total_businesses"] = totalBusinesses

	// Active businesses
	var activeBusinesses int64
	if err := r.db.Model(&models.Business{}).Where("status = 1").Count(&activeBusinesses).Error; err != nil {
		return nil, err
	}
	stats["active_businesses"] = activeBusinesses

	// Inactive businesses
	var inactiveBusinesses int64
	if err := r.db.Model(&models.Business{}).Where("status = 0").Count(&inactiveBusinesses).Error; err != nil {
		return nil, err
	}
	stats["inactive_businesses"] = inactiveBusinesses

	// Businesses with packages
	var businessesWithPackages int64
	if err := r.db.Model(&models.Business{}).Where("package_id IS NOT NULL").Count(&businessesWithPackages).Error; err != nil {
		return nil, err
	}
	stats["businesses_with_packages"] = businessesWithPackages

	// Businesses without packages
	var businessesWithoutPackages int64
	if err := r.db.Model(&models.Business{}).Where("package_id IS NULL").Count(&businessesWithoutPackages).Error; err != nil {
		return nil, err
	}
	stats["businesses_without_packages"] = businessesWithoutPackages

	return stats, nil
}

func (r *businessRepository) GetLocationStats() (map[string]int64, error) {
	type LocationStat struct {
		Location string `json:"location"`
		Count    int64  `json:"count"`
	}

	var stats []LocationStat
	err := r.db.Model(&models.Business{}).
		Select("location, COUNT(*) as count").
		Where("location IS NOT NULL AND location != ''").
		Group("location").
		Order("count DESC").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.Location] = stat.Count
	}

	return result, nil
}

func (r *businessRepository) GetPackageDistribution() (map[string]int64, error) {
	type PackageDistribution struct {
		PackageName string `json:"package_name"`
		Count       int64  `json:"count"`
	}

	var stats []PackageDistribution
	err := r.db.Model(&models.Business{}).
		Select("COALESCE(packages.name, 'No Package') as package_name, COUNT(*) as count").
		Joins("LEFT JOIN packages ON businesses.package_id = packages.id").
		Group("packages.name").
		Order("count DESC").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.PackageName] = stat.Count
	}

	return result, nil
}

// Relationships

func (r *businessRepository) GetBusinessWithRelations(id uint) (*models.Business, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid business ID")
	}

	var business models.Business
	err := r.db.Preload("User").Preload("Package").First(&business, id).Error
	if err != nil {
		return nil, err
	}
	return &business, nil
}

func (r *businessRepository) GetBySlugWithRelations(slug string) (*models.Business, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	var business models.Business
	err := r.db.Preload("User").Preload("Package").Where("slug = ?", slug).First(&business).Error
	if err != nil {
		return nil, err
	}
	return &business, nil
}

// Bulk operations

func (r *businessRepository) BulkUpdateStatus(businessIDs []uint, status int) error {
	if len(businessIDs) == 0 {
		return fmt.Errorf("no business IDs provided")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Business{}).
		Where("id IN ?", businessIDs).
		Update("status", status).Error
}

func (r *businessRepository) BulkAssignPackage(businessIDs []uint, packageID uint) error {
	if len(businessIDs) == 0 {
		return fmt.Errorf("no business IDs provided")
	}
	if packageID == 0 {
		return fmt.Errorf("invalid package ID")
	}

	return r.db.Model(&models.Business{}).
		Where("id IN ?", businessIDs).
		Update("package_id", packageID).Error
}

// Location operations

func (r *businessRepository) GetBusinessesByLocation(location string) ([]models.Business, error) {
	if location == "" {
		return nil, fmt.Errorf("location cannot be empty")
	}

	var businesses []models.Business
	err := r.db.Where("location ILIKE ?", "%"+location+"%").Find(&businesses).Error
	return businesses, err
}

func (r *businessRepository) GetBusinessLocations() ([]string, error) {
	var locations []string
	err := r.db.Model(&models.Business{}).
		Select("DISTINCT location").
		Where("location IS NOT NULL AND location != ''").
		Order("location").
		Pluck("location", &locations).Error
	return locations, err
}

// Transaction support

func (r *businessRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}
