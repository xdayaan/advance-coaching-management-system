package repository

import (
	"backend/internal/models"
	"backend/pkg/database"
	"fmt"

	"gorm.io/gorm"
)

type StudentRepository interface {
	// Basic CRUD operations
	Create(student *models.Student) error
	CreateWithTransaction(tx *gorm.DB, student *models.Student) error
	GetByID(id uint) (*models.Student, error)
	GetByUserID(userID uint) (*models.Student, error)
	GetAll(filters StudentFilters) ([]models.Student, int64, error)
	GetAllWithRelations(filters StudentFilters) ([]models.Student, int64, error)
	Update(student *models.Student) error
	UpdateWithTransaction(tx *gorm.DB, student *models.Student) error
	Delete(id uint) error

	// Business specific operations
	GetByBusinessID(businessID uint, filters StudentFilters) ([]models.Student, int64, error)
	GetActiveStudentsByBusiness(businessID uint) ([]models.Student, error)
	GetInactiveStudentsByBusiness(businessID uint) ([]models.Student, error)

	// Status operations
	UpdateStudentStatus(studentID uint, status int) error
	GetActiveStudents() ([]models.Student, error)
	GetInactiveStudents() ([]models.Student, error)

	// Search and filters
	SearchStudents(searchTerm string, limit int, businessID ...uint) ([]models.Student, error)
	SearchStudentsByBusiness(businessID uint, searchTerm string, limit int) ([]models.Student, error)

	// Statistics
	GetStudentStats(businessID ...uint) (map[string]interface{}, error)
	GetGuardianStats(businessID ...uint) (map[string]interface{}, error)

	// Relationships
	GetStudentWithRelations(id uint) (*models.Student, error)

	// Bulk operations
	BulkUpdateStatus(studentIDs []uint, status int) error

	// Validation
	StudentUserExists(userID uint, excludeStudentID ...uint) (bool, error)

	// Transaction support
	BeginTransaction() *gorm.DB
}

type StudentFilters struct {
	BusinessID    *uint  `form:"business_id" json:"business_id"`
	Status        *int   `form:"status" json:"status"`
	GuardianName  string `form:"guardian_name" json:"guardian_name"`
	GuardianEmail string `form:"guardian_email" json:"guardian_email"`
	Search        string `form:"search" json:"search"`
	Page          int    `form:"page" json:"page"`
	Limit         int    `form:"limit" json:"limit"`
	SortBy        string `form:"sort_by" json:"sort_by"`
	SortOrder     string `form:"sort_order" json:"sort_order"`
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository() StudentRepository {
	return &studentRepository{
		db: database.DB,
	}
}

func (r *studentRepository) Create(student *models.Student) error {
	if student == nil {
		return fmt.Errorf("student cannot be nil")
	}
	return r.db.Create(student).Error
}

func (r *studentRepository) CreateWithTransaction(tx *gorm.DB, student *models.Student) error {
	if student == nil {
		return fmt.Errorf("student cannot be nil")
	}
	return tx.Create(student).Error
}

func (r *studentRepository) GetByID(id uint) (*models.Student, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid student ID")
	}

	var student models.Student
	err := r.db.First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) GetByUserID(userID uint) (*models.Student, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	var student models.Student
	err := r.db.Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) GetAll(filters StudentFilters) ([]models.Student, int64, error) {
	var students []models.Student
	var total int64

	query := r.db.Model(&models.Student{})

	// Apply filters
	if filters.BusinessID != nil {
		query = query.Where("business_id = ?", *filters.BusinessID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.GuardianName != "" {
		query = query.Where("guardian_name ILIKE ?", "%"+filters.GuardianName+"%")
	}

	if filters.GuardianEmail != "" {
		query = query.Where("guardian_email ILIKE ?", "%"+filters.GuardianEmail+"%")
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR guardian_name ILIKE ? OR guardian_email ILIKE ? OR guardian_number ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_on DESC"
	if filters.SortBy != "" {
		validSortFields := map[string]bool{
			"created_on":      true,
			"updated_on":      true,
			"name":            true,
			"guardian_name":   true,
			"guardian_email":  true,
			"guardian_number": true,
			"status":          true,
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

	err := query.Find(&students).Error
	return students, total, err
}

func (r *studentRepository) GetAllWithRelations(filters StudentFilters) ([]models.Student, int64, error) {
	var students []models.Student
	var total int64

	query := r.db.Model(&models.Student{}).Preload("User").Preload("Business")

	// Apply same filters as GetAll
	if filters.BusinessID != nil {
		query = query.Where("business_id = ?", *filters.BusinessID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.GuardianName != "" {
		query = query.Where("guardian_name ILIKE ?", "%"+filters.GuardianName+"%")
	}

	if filters.GuardianEmail != "" {
		query = query.Where("guardian_email ILIKE ?", "%"+filters.GuardianEmail+"%")
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR guardian_name ILIKE ? OR guardian_email ILIKE ? OR guardian_number ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_on DESC"
	if filters.SortBy != "" {
		validSortFields := map[string]bool{
			"created_on":      true,
			"updated_on":      true,
			"name":            true,
			"guardian_name":   true,
			"guardian_email":  true,
			"guardian_number": true,
			"status":          true,
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

	err := query.Find(&students).Error
	return students, total, err
}

func (r *studentRepository) Update(student *models.Student) error {
	if student == nil {
		return fmt.Errorf("student cannot be nil")
	}
	if student.ID == 0 {
		return fmt.Errorf("student ID cannot be zero")
	}
	return r.db.Save(student).Error
}

func (r *studentRepository) UpdateWithTransaction(tx *gorm.DB, student *models.Student) error {
	if student == nil {
		return fmt.Errorf("student cannot be nil")
	}
	if student.ID == 0 {
		return fmt.Errorf("student ID cannot be zero")
	}
	return tx.Save(student).Error
}

func (r *studentRepository) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid student ID")
	}
	return r.db.Delete(&models.Student{}, id).Error
}

func (r *studentRepository) GetByBusinessID(businessID uint, filters StudentFilters) ([]models.Student, int64, error) {
	if businessID == 0 {
		return nil, 0, fmt.Errorf("invalid business ID")
	}

	filters.BusinessID = &businessID
	return r.GetAll(filters)
}

func (r *studentRepository) GetActiveStudentsByBusiness(businessID uint) ([]models.Student, error) {
	if businessID == 0 {
		return nil, fmt.Errorf("invalid business ID")
	}

	var students []models.Student
	err := r.db.Where("business_id = ? AND status = 1", businessID).Find(&students).Error
	return students, err
}

func (r *studentRepository) GetInactiveStudentsByBusiness(businessID uint) ([]models.Student, error) {
	if businessID == 0 {
		return nil, fmt.Errorf("invalid business ID")
	}

	var students []models.Student
	err := r.db.Where("business_id = ? AND status = 0", businessID).Find(&students).Error
	return students, err
}

func (r *studentRepository) UpdateStudentStatus(studentID uint, status int) error {
	if studentID == 0 {
		return fmt.Errorf("invalid student ID")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Student{}).Where("id = ?", studentID).Update("status", status).Error
}

func (r *studentRepository) GetActiveStudents() ([]models.Student, error) {
	var students []models.Student
	err := r.db.Where("status = 1").Find(&students).Error
	return students, err
}

func (r *studentRepository) GetInactiveStudents() ([]models.Student, error) {
	var students []models.Student
	err := r.db.Where("status = 0").Find(&students).Error
	return students, err
}

func (r *studentRepository) SearchStudents(searchTerm string, limit int, businessID ...uint) ([]models.Student, error) {
	if searchTerm == "" {
		return []models.Student{}, nil
	}

	query := r.db.Where("name ILIKE ? OR guardian_name ILIKE ? OR guardian_email ILIKE ? OR guardian_number ILIKE ?",
		"%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%")

	if len(businessID) > 0 && businessID[0] > 0 {
		query = query.Where("business_id = ?", businessID[0])
	}

	query = query.Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var students []models.Student
	err := query.Find(&students).Error
	return students, err
}

func (r *studentRepository) SearchStudentsByBusiness(businessID uint, searchTerm string, limit int) ([]models.Student, error) {
	return r.SearchStudents(searchTerm, limit, businessID)
}

func (r *studentRepository) GetStudentStats(businessID ...uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.Model(&models.Student{})
	if len(businessID) > 0 && businessID[0] > 0 {
		query = query.Where("business_id = ?", businessID[0])
	}

	// Total students
	var totalStudents int64
	if err := query.Count(&totalStudents).Error; err != nil {
		return nil, err
	}
	stats["total_students"] = totalStudents

	// Active students
	var activeStudents int64
	activeQuery := query
	if err := activeQuery.Where("status = 1").Count(&activeStudents).Error; err != nil {
		return nil, err
	}
	stats["active_students"] = activeStudents

	// Inactive students
	stats["inactive_students"] = totalStudents - activeStudents

	return stats, nil
}

func (r *studentRepository) GetGuardianStats(businessID ...uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.Model(&models.Student{})
	if len(businessID) > 0 && businessID[0] > 0 {
		query = query.Where("business_id = ?", businessID[0])
	}

	// Students with guardian email
	var studentsWithEmail int64
	if err := query.Where("guardian_email IS NOT NULL AND guardian_email != ''").Count(&studentsWithEmail).Error; err != nil {
		return nil, err
	}
	stats["students_with_guardian_email"] = studentsWithEmail

	// Students with guardian phone
	var studentsWithPhone int64
	if err := query.Where("guardian_number IS NOT NULL AND guardian_number != ''").Count(&studentsWithPhone).Error; err != nil {
		return nil, err
	}
	stats["students_with_guardian_phone"] = studentsWithPhone

	return stats, nil
}

func (r *studentRepository) GetStudentWithRelations(id uint) (*models.Student, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid student ID")
	}

	var student models.Student
	err := r.db.Preload("User").Preload("Business").First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) BulkUpdateStatus(studentIDs []uint, status int) error {
	if len(studentIDs) == 0 {
		return fmt.Errorf("no student IDs provided")
	}
	if status < 0 || status > 1 {
		return fmt.Errorf("invalid status value")
	}

	return r.db.Model(&models.Student{}).
		Where("id IN ?", studentIDs).
		Update("status", status).Error
}

func (r *studentRepository) StudentUserExists(userID uint, excludeStudentID ...uint) (bool, error) {
	if userID == 0 {
		return false, fmt.Errorf("user ID cannot be zero")
	}

	var count int64
	query := r.db.Model(&models.Student{}).Where("user_id = ?", userID)

	if len(excludeStudentID) > 0 && excludeStudentID[0] > 0 {
		query = query.Where("id != ?", excludeStudentID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *studentRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}
