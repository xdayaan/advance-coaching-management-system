package repository

import (
	"backend/internal/models"
	"backend/pkg/database"
	"fmt"

	"gorm.io/gorm"
)

type TeacherRepository interface {
	// Basic CRUD operations
	Create(teacher *models.Teacher) error
	CreateWithTransaction(tx *gorm.DB, teacher *models.Teacher) error
	GetByID(id uint) (*models.Teacher, error)
	GetByUserID(userID uint) (*models.Teacher, error)
	GetAll(filters TeacherFilters) ([]models.Teacher, int64, error)
	GetAllWithRelations(filters TeacherFilters) ([]models.Teacher, int64, error)
	Update(teacher *models.Teacher) error
	UpdateWithTransaction(tx *gorm.DB, teacher *models.Teacher) error
	Delete(id uint) error

	// Business specific operations
	GetByBusinessID(businessID uint, filters TeacherFilters) ([]models.Teacher, int64, error)
	GetActiveTeachersByBusiness(businessID uint) ([]models.Teacher, error)
	GetInactiveTeachersByBusiness(businessID uint) ([]models.Teacher, error)

	// Status operations
	UpdateTeacherStatus(teacherID uint, status int) error
	GetActiveTeachers() ([]models.Teacher, error)
	GetInactiveTeachers() ([]models.Teacher, error)

	// Search and filters
	SearchTeachers(searchTerm string, limit int, businessID ...uint) ([]models.Teacher, error)
	SearchTeachersByBusiness(businessID uint, searchTerm string, limit int) ([]models.Teacher, error)

	// Statistics
	GetTeacherStats(businessID ...uint) (map[string]interface{}, error)
	GetSalaryStats(businessID ...uint) (map[string]interface{}, error)
	GetQualificationStats(businessID ...uint) (map[string]int64, error)

	// Relationships
	GetTeacherWithRelations(id uint) (*models.Teacher, error)

	// Bulk operations
	BulkUpdateStatus(teacherIDs []uint, status int) error
	BulkUpdateSalary(teacherIDs []uint, salary float64) error

	// Validation
	TeacherUserExists(userID uint, excludeTeacherID ...uint) (bool, error)

	// Transaction support
	BeginTransaction() *gorm.DB
}

type TeacherFilters struct {
	BusinessID    *uint    `form:"business_id" json:"business_id"`
	Status        *int     `form:"status" json:"status"`
	MinSalary     *float64 `form:"min_salary" json:"min_salary"`
	MaxSalary     *float64 `form:"max_salary" json:"max_salary"`
	Qualification string   `form:"qualification" json:"qualification"`
	Search        string   `form:"search" json:"search"`
	Page          int      `form:"page" json:"page"`
	Limit         int      `form:"limit" json:"limit"`
	SortBy        string   `form:"sort_by" json:"sort_by"`
	SortOrder     string   `form:"sort_order" json:"sort_order"`
}

type teacherRepository struct {
	db *gorm.DB
}

func NewTeacherRepository() TeacherRepository {
	return &teacherRepository{
		db: database.DB,
	}
}

func (r *teacherRepository) Create(teacher *models.Teacher) error {
	if teacher == nil {
		return fmt.Errorf("teacher cannot be nil")
	}
	return r.db.Create(teacher).Error
}

func (r *teacherRepository) CreateWithTransaction(tx *gorm.DB, teacher *models.Teacher) error {
	if teacher == nil {
		return fmt.Errorf("teacher cannot be nil")
	}
	return tx.Create(teacher).Error
}

func (r *teacherRepository) GetByID(id uint) (*models.Teacher, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid teacher ID")
	}

	var teacher models.Teacher
	err := r.db.First(&teacher, id).Error
	if err != nil {
		return nil, err
	}
	return &teacher, nil
}

func (r *teacherRepository) GetByUserID(userID uint) (*models.Teacher, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	var teacher models.Teacher
	err := r.db.Where("user_id = ?", userID).First(&teacher).Error
	if err != nil {
		return nil, err
	}
	return &teacher, nil
}

func (r *teacherRepository) GetAll(filters TeacherFilters) ([]models.Teacher, int64, error) {
	var teachers []models.Teacher
	var total int64

	query := r.db.Model(&models.Teacher{})

	// Apply filters
	if filters.BusinessID != nil {
		query = query.Where("business_id = ?", *filters.BusinessID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.MinSalary != nil {
		query = query.Where("salary >= ?", *filters.MinSalary)
	}

	if filters.MaxSalary != nil {
		query = query.Where("salary <= ?", *filters.MaxSalary)
	}

	if filters.Qualification != "" {
		query = query.Where("qualification ILIKE ?", "%"+filters.Qualification+"%")
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR qualification ILIKE ? OR experience ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_on DESC"
	if filters.SortBy != "" {
		validSortFields := map[string]bool{
			"created_on":    true,
			"updated_on":    true,
			"name":          true,
			"salary":        true,
			"qualification": true,
			"experience":    true,
			"status":        true,
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

	err := query.Find(&teachers).Error
	return teachers, total, err
}

func (r *teacherRepository) GetAllWithRelations(filters TeacherFilters) ([]models.Teacher, int64, error) {
	var teachers []models.Teacher
	var total int64

	query := r.db.Model(&models.Teacher{}).Preload("User").Preload("Business")

	// Apply same filters as GetAll
	if filters.BusinessID != nil {
		query = query.Where("business_id = ?", *filters.BusinessID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.MinSalary != nil {
		query = query.Where("salary >= ?", *filters.MinSalary)
	}

	if filters.MaxSalary != nil {
		query = query.Where("salary <= ?", *filters.MaxSalary)
	}

	if filters.Qualification != "" {
		query = query.Where("qualification ILIKE ?", "%"+filters.Qualification+"%")
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR qualification ILIKE ? OR experience ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_on DESC"
	if filters.SortBy != "" {
		validSortFields := map[string]bool{
			"created_on":    true,
			"updated_on":    true,
			"name":          true,
			"salary":        true,
			"qualification": true,
			"experience":    true,
			"status":        true,
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

	err := query.Find(&teachers).Error
	return teachers, total, err
}

func (r *teacherRepository) Update(teacher *models.Teacher) error {
	if teacher == nil {
		return fmt.Errorf("teacher cannot be nil")
	}
	if teacher.ID == 0 {
		return fmt.Errorf("teacher ID cannot be zero")
	}
	return r.db.Save(teacher).Error
}

func (r *teacherRepository) UpdateWithTransaction(tx *gorm.DB, teacher *models.Teacher) error {
	if teacher == nil {
		return fmt.Errorf("teacher cannot be nil")
	}
	if teacher.ID == 0 {
		return fmt.Errorf("teacher ID cannot be zero")
	}
	return tx.Save(teacher).Error
}

func (r *teacherRepository) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid teacher ID")
	}
	return r.db.Delete(&models.Teacher{}, id).Error
}

func (r *teacherRepository) GetByBusinessID(businessID uint, filters TeacherFilters) ([]models.Teacher, int64, error) {
	if businessID == 0 {
		return nil, 0, fmt.Errorf("invalid business ID")
	}

	filters.BusinessID = &businessID
	return r.GetAll(filters)
}

func (r *teacherRepository) GetActiveTeachersByBusiness(businessID uint) ([]models.Teacher, error) {
	if businessID == 0 {
		return nil, fmt.Errorf("invalid business ID")
	}

	var teachers []models.Teacher
	err := r.db.Where("business_id = ? AND status = 1", businessID).Find(&teachers).Error
	return teachers, err
}

func (r *teacherRepository) GetInactiveTeachersByBusiness(businessID uint) ([]models.Teacher, error) {
	if businessID == 0 {
		return nil, fmt.Errorf("invalid business ID")
	}

	var teachers []models.Teacher
	err := r.db.Where("business_id = ? AND status = 0", businessID).Find(&teachers).Error
	return teachers, err
}

func (r *teacherRepository) UpdateTeacherStatus(teacherID uint, status int) error {
	if teacherID == 0 {
		return fmt.Errorf("invalid teacher ID")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Teacher{}).Where("id = ?", teacherID).Update("status", status).Error
}

func (r *teacherRepository) GetActiveTeachers() ([]models.Teacher, error) {
	var teachers []models.Teacher
	err := r.db.Where("status = 1").Find(&teachers).Error
	return teachers, err
}

func (r *teacherRepository) GetInactiveTeachers() ([]models.Teacher, error) {
	var teachers []models.Teacher
	err := r.db.Where("status = 0").Find(&teachers).Error
	return teachers, err
}

func (r *teacherRepository) SearchTeachers(searchTerm string, limit int, businessID ...uint) ([]models.Teacher, error) {
	if searchTerm == "" {
		return []models.Teacher{}, nil
	}

	query := r.db.Where("name ILIKE ? OR qualification ILIKE ? OR experience ILIKE ?",
		"%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%")

	if len(businessID) > 0 && businessID[0] > 0 {
		query = query.Where("business_id = ?", businessID[0])
	}

	query = query.Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var teachers []models.Teacher
	err := query.Find(&teachers).Error
	return teachers, err
}

func (r *teacherRepository) SearchTeachersByBusiness(businessID uint, searchTerm string, limit int) ([]models.Teacher, error) {
	return r.SearchTeachers(searchTerm, limit, businessID)
}

func (r *teacherRepository) GetTeacherStats(businessID ...uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.Model(&models.Teacher{})
	if len(businessID) > 0 && businessID[0] > 0 {
		query = query.Where("business_id = ?", businessID[0])
	}

	// Total teachers
	var totalTeachers int64
	if err := query.Count(&totalTeachers).Error; err != nil {
		return nil, err
	}
	stats["total_teachers"] = totalTeachers

	// Active teachers
	var activeTeachers int64
	activeQuery := query
	if err := activeQuery.Where("status = 1").Count(&activeTeachers).Error; err != nil {
		return nil, err
	}
	stats["active_teachers"] = activeTeachers

	// Inactive teachers
	stats["inactive_teachers"] = totalTeachers - activeTeachers

	return stats, nil
}

func (r *teacherRepository) GetSalaryStats(businessID ...uint) (map[string]interface{}, error) {
	type SalaryStats struct {
		MinSalary float64 `json:"min_salary"`
		MaxSalary float64 `json:"max_salary"`
		AvgSalary float64 `json:"avg_salary"`
	}

	query := r.db.Model(&models.Teacher{})
	if len(businessID) > 0 && businessID[0] > 0 {
		query = query.Where("business_id = ?", businessID[0])
	}

	var stats SalaryStats
	err := query.Select("MIN(salary) as min_salary, MAX(salary) as max_salary, AVG(salary) as avg_salary").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["min_salary"] = stats.MinSalary
	result["max_salary"] = stats.MaxSalary
	result["avg_salary"] = stats.AvgSalary

	return result, nil
}

func (r *teacherRepository) GetQualificationStats(businessID ...uint) (map[string]int64, error) {
	type QualificationStat struct {
		Qualification string `json:"qualification"`
		Count         int64  `json:"count"`
	}

	query := r.db.Model(&models.Teacher{})
	if len(businessID) > 0 && businessID[0] > 0 {
		query = query.Where("business_id = ?", businessID[0])
	}

	var stats []QualificationStat
	err := query.Select("qualification, COUNT(*) as count").
		Where("qualification IS NOT NULL AND qualification != ''").
		Group("qualification").
		Order("count DESC").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.Qualification] = stat.Count
	}

	return result, nil
}

func (r *teacherRepository) GetTeacherWithRelations(id uint) (*models.Teacher, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid teacher ID")
	}

	var teacher models.Teacher
	err := r.db.Preload("User").Preload("Business").First(&teacher, id).Error
	if err != nil {
		return nil, err
	}
	return &teacher, nil
}

func (r *teacherRepository) BulkUpdateStatus(teacherIDs []uint, status int) error {
	if len(teacherIDs) == 0 {
		return fmt.Errorf("no teacher IDs provided")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Teacher{}).
		Where("id IN ?", teacherIDs).
		Update("status", status).Error
}

func (r *teacherRepository) BulkUpdateSalary(teacherIDs []uint, salary float64) error {
	if len(teacherIDs) == 0 {
		return fmt.Errorf("no teacher IDs provided")
	}
	if salary < 0 {
		return fmt.Errorf("invalid salary value")
	}

	return r.db.Model(&models.Teacher{}).
		Where("id IN ?", teacherIDs).
		Update("salary", salary).Error
}

func (r *teacherRepository) TeacherUserExists(userID uint, excludeTeacherID ...uint) (bool, error) {
	if userID == 0 {
		return false, fmt.Errorf("user ID cannot be zero")
	}

	var count int64
	query := r.db.Model(&models.Teacher{}).Where("user_id = ?", userID)

	if len(excludeTeacherID) > 0 && excludeTeacherID[0] > 0 {
		query = query.Where("id != ?", excludeTeacherID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *teacherRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}
