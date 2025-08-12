package services

import (
	"backend/internal/models"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type TeacherService interface {
	// Basic CRUD operations
	CreateTeacher(req models.CreateTeacherRequest) (*models.TeacherResponse, error)
	GetTeacherByID(id uint) (*models.TeacherResponse, error)
	GetTeacherByUserID(userID uint) (*models.TeacherResponse, error)
	GetTeachers(filters repository.TeacherFilters) ([]models.TeacherResponse, int64, error)
	UpdateTeacher(teacherID uint, updates map[string]interface{}) (*models.TeacherResponse, error)
	DeleteTeacher(teacherID uint) error

	// Business specific operations
	GetTeachersByBusiness(businessID uint, filters repository.TeacherFilters) ([]models.TeacherResponse, int64, error)
	GetActiveTeachersByBusiness(businessID uint) ([]models.TeacherResponse, error)
	GetInactiveTeachersByBusiness(businessID uint) ([]models.TeacherResponse, error)

	// Status operations
	ChangeTeacherStatus(teacherID uint, status int) error
	GetActiveTeachers() ([]models.TeacherResponse, error)
	GetInactiveTeachers() ([]models.TeacherResponse, error)

	// Search
	SearchTeachers(searchTerm string, limit int, businessID ...uint) ([]models.TeacherResponse, error)
	SearchTeachersByBusiness(businessID uint, searchTerm string, limit int) ([]models.TeacherResponse, error)

	// Statistics
	GetTeacherStats(businessID ...uint) (map[string]interface{}, error)
	GetSalaryStats(businessID ...uint) (map[string]interface{}, error)
	GetQualificationStats(businessID ...uint) (map[string]int64, error)

	// Bulk operations
	BulkUpdateTeacherStatus(teacherIDs []uint, status int) error
	BulkUpdateSalary(teacherIDs []uint, salary float64) error

	// Validation
	ValidateCreateTeacherRequest(req models.CreateTeacherRequest) error
	ValidateUpdateTeacherRequest(req models.UpdateTeacherRequest) error
}

type teacherService struct {
	teacherRepo  repository.TeacherRepository
	userRepo     repository.UserRepository
	businessRepo repository.BusinessRepository
}

func NewTeacherService(teacherRepo repository.TeacherRepository, userRepo repository.UserRepository, businessRepo repository.BusinessRepository) TeacherService {
	return &teacherService{
		teacherRepo:  teacherRepo,
		userRepo:     userRepo,
		businessRepo: businessRepo,
	}
}

func (s *teacherService) CreateTeacher(req models.CreateTeacherRequest) (*models.TeacherResponse, error) {
	// Validate request
	if err := s.ValidateCreateTeacherRequest(req); err != nil {
		return nil, err
	}

	// Check if user exists and is not already a teacher
	_, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is already a teacher
	exists, err := s.teacherRepo.TeacherUserExists(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user is already a teacher")
	}
	if exists {
		return nil, fmt.Errorf("user is already a teacher")
	}

	// Check if business exists
	_, err = s.businessRepo.GetByID(req.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("business not found")
	}

	// Create teacher
	teacher := &models.Teacher{
		Name:          req.Name,
		UserID:        req.UserID,
		BusinessID:    req.BusinessID,
		Salary:        req.Salary,
		Qualification: req.Qualification,
		Experience:    req.Experience,
		Description:   req.Description,
		Status:        1, // Active by default
	}

	if err := s.teacherRepo.Create(teacher); err != nil {
		return nil, fmt.Errorf("failed to create teacher: %v", err)
	}

	// Get teacher with relations
	teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created teacher")
	}

	return s.toTeacherResponse(teacherWithRelations), nil
}

func (s *teacherService) GetTeacherByID(id uint) (*models.TeacherResponse, error) {
	teacher, err := s.teacherRepo.GetTeacherWithRelations(id)
	if err != nil {
		return nil, fmt.Errorf("teacher not found")
	}

	return s.toTeacherResponse(teacher), nil
}

func (s *teacherService) GetTeacherByUserID(userID uint) (*models.TeacherResponse, error) {
	teacher, err := s.teacherRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("teacher profile not found")
	}

	// Get with relations
	teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher details")
	}

	return s.toTeacherResponse(teacherWithRelations), nil
}

func (s *teacherService) GetTeachers(filters repository.TeacherFilters) ([]models.TeacherResponse, int64, error) {
	// Set default pagination
	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Limit == 0 {
		filters.Limit = 10
	}

	teachers, total, err := s.teacherRepo.GetAllWithRelations(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get teachers: %v", err)
	}

	var responses []models.TeacherResponse
	for _, teacher := range teachers {
		responses = append(responses, *s.toTeacherResponse(&teacher))
	}

	return responses, total, nil
}

func (s *teacherService) UpdateTeacher(teacherID uint, updates map[string]interface{}) (*models.TeacherResponse, error) {
	// Get existing teacher
	teacher, err := s.teacherRepo.GetByID(teacherID)
	if err != nil {
		return nil, fmt.Errorf("teacher not found")
	}

	// Update fields
	if name, ok := updates["name"]; ok {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			teacher.Name = nameStr
		}
	}

	if salary, ok := updates["salary"]; ok {
		if salaryFloat, ok := salary.(float64); ok && salaryFloat >= 0 {
			teacher.Salary = salaryFloat
		}
	}

	if qualification, ok := updates["qualification"]; ok {
		if qualStr, ok := qualification.(string); ok {
			teacher.Qualification = qualStr
		}
	}

	if experience, ok := updates["experience"]; ok {
		if expStr, ok := experience.(string); ok {
			teacher.Experience = expStr
		}
	}

	if description, ok := updates["description"]; ok {
		if descStr, ok := description.(string); ok {
			teacher.Description = descStr
		}
	}

	if status, ok := updates["status"]; ok {
		if statusInt, ok := status.(int); ok && (statusInt == 0 || statusInt == 1) {
			teacher.Status = statusInt
		}
	}

	// Save updates
	if err := s.teacherRepo.Update(teacher); err != nil {
		return nil, fmt.Errorf("failed to update teacher: %v", err)
	}

	// Get updated teacher with relations
	updatedTeacher, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated teacher")
	}

	return s.toTeacherResponse(updatedTeacher), nil
}

func (s *teacherService) DeleteTeacher(teacherID uint) error {
	// Check if teacher exists
	_, err := s.teacherRepo.GetByID(teacherID)
	if err != nil {
		return fmt.Errorf("teacher not found")
	}

	if err := s.teacherRepo.Delete(teacherID); err != nil {
		return fmt.Errorf("failed to delete teacher: %v", err)
	}

	return nil
}

func (s *teacherService) GetTeachersByBusiness(businessID uint, filters repository.TeacherFilters) ([]models.TeacherResponse, int64, error) {
	// Set default pagination
	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Limit == 0 {
		filters.Limit = 10
	}

	teachers, total, err := s.teacherRepo.GetByBusinessID(businessID, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get teachers by business: %v", err)
	}

	var responses []models.TeacherResponse
	for _, teacher := range teachers {
		// Get with relations for complete data
		teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
		if err != nil {
			continue // Skip this one if there's an issue
		}
		responses = append(responses, *s.toTeacherResponse(teacherWithRelations))
	}

	return responses, total, nil
}

func (s *teacherService) GetActiveTeachersByBusiness(businessID uint) ([]models.TeacherResponse, error) {
	teachers, err := s.teacherRepo.GetActiveTeachersByBusiness(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active teachers: %v", err)
	}

	var responses []models.TeacherResponse
	for _, teacher := range teachers {
		teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toTeacherResponse(teacherWithRelations))
	}

	return responses, nil
}

func (s *teacherService) GetInactiveTeachersByBusiness(businessID uint) ([]models.TeacherResponse, error) {
	teachers, err := s.teacherRepo.GetInactiveTeachersByBusiness(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive teachers: %v", err)
	}

	var responses []models.TeacherResponse
	for _, teacher := range teachers {
		teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toTeacherResponse(teacherWithRelations))
	}

	return responses, nil
}

func (s *teacherService) ChangeTeacherStatus(teacherID uint, status int) error {
	// Check if teacher exists
	_, err := s.teacherRepo.GetByID(teacherID)
	if err != nil {
		return fmt.Errorf("teacher not found")
	}

	if err := s.teacherRepo.UpdateTeacherStatus(teacherID, status); err != nil {
		return fmt.Errorf("failed to update teacher status: %v", err)
	}

	return nil
}

func (s *teacherService) GetActiveTeachers() ([]models.TeacherResponse, error) {
	teachers, err := s.teacherRepo.GetActiveTeachers()
	if err != nil {
		return nil, fmt.Errorf("failed to get active teachers: %v", err)
	}

	var responses []models.TeacherResponse
	for _, teacher := range teachers {
		teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toTeacherResponse(teacherWithRelations))
	}

	return responses, nil
}

func (s *teacherService) GetInactiveTeachers() ([]models.TeacherResponse, error) {
	teachers, err := s.teacherRepo.GetInactiveTeachers()
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive teachers: %v", err)
	}

	var responses []models.TeacherResponse
	for _, teacher := range teachers {
		teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toTeacherResponse(teacherWithRelations))
	}

	return responses, nil
}

func (s *teacherService) SearchTeachers(searchTerm string, limit int, businessID ...uint) ([]models.TeacherResponse, error) {
	teachers, err := s.teacherRepo.SearchTeachers(searchTerm, limit, businessID...)
	if err != nil {
		return nil, fmt.Errorf("failed to search teachers: %v", err)
	}

	var responses []models.TeacherResponse
	for _, teacher := range teachers {
		teacherWithRelations, err := s.teacherRepo.GetTeacherWithRelations(teacher.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toTeacherResponse(teacherWithRelations))
	}

	return responses, nil
}

func (s *teacherService) SearchTeachersByBusiness(businessID uint, searchTerm string, limit int) ([]models.TeacherResponse, error) {
	return s.SearchTeachers(searchTerm, limit, businessID)
}

func (s *teacherService) GetTeacherStats(businessID ...uint) (map[string]interface{}, error) {
	return s.teacherRepo.GetTeacherStats(businessID...)
}

func (s *teacherService) GetSalaryStats(businessID ...uint) (map[string]interface{}, error) {
	return s.teacherRepo.GetSalaryStats(businessID...)
}

func (s *teacherService) GetQualificationStats(businessID ...uint) (map[string]int64, error) {
	return s.teacherRepo.GetQualificationStats(businessID...)
}

func (s *teacherService) BulkUpdateTeacherStatus(teacherIDs []uint, status int) error {
	if len(teacherIDs) == 0 {
		return fmt.Errorf("no teacher IDs provided")
	}

	if err := s.teacherRepo.BulkUpdateStatus(teacherIDs, status); err != nil {
		return fmt.Errorf("failed to bulk update teacher status: %v", err)
	}

	return nil
}

func (s *teacherService) BulkUpdateSalary(teacherIDs []uint, salary float64) error {
	if len(teacherIDs) == 0 {
		return fmt.Errorf("no teacher IDs provided")
	}

	if salary < 0 {
		return fmt.Errorf("salary cannot be negative")
	}

	if err := s.teacherRepo.BulkUpdateSalary(teacherIDs, salary); err != nil {
		return fmt.Errorf("failed to bulk update teacher salary: %v", err)
	}

	return nil
}

func (s *teacherService) ValidateCreateTeacherRequest(req models.CreateTeacherRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if req.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}

	if req.BusinessID == 0 {
		return fmt.Errorf("business ID is required")
	}

	if req.Salary < 0 {
		return fmt.Errorf("salary cannot be negative")
	}

	return nil
}

func (s *teacherService) ValidateUpdateTeacherRequest(req models.UpdateTeacherRequest) error {
	if req.Name != "" && strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if req.Salary != nil && *req.Salary < 0 {
		return fmt.Errorf("salary cannot be negative")
	}

	if req.Status != nil && (*req.Status < 0 || *req.Status > 1) {
		return fmt.Errorf("invalid status value")
	}

	return nil
}

// Helper methods
func (s *teacherService) toTeacherResponse(teacher *models.Teacher) *models.TeacherResponse {
	response := &models.TeacherResponse{
		ID:            teacher.ID,
		Name:          teacher.Name,
		UserID:        teacher.UserID,
		BusinessID:    teacher.BusinessID,
		Salary:        teacher.Salary,
		Qualification: teacher.Qualification,
		Experience:    teacher.Experience,
		Description:   teacher.Description,
		Status:        teacher.Status,
		CreatedOn:     teacher.CreatedOn,
		UpdatedOn:     teacher.UpdatedOn,
	}

	// Add user details if loaded
	if teacher.User.ID != 0 {
		response.User = &models.UserResponse{
			ID:        teacher.User.ID,
			Name:      teacher.User.Name,
			Email:     teacher.User.Email,
			Role:      teacher.User.Role,
			Status:    teacher.User.Status,
			CreatedOn: teacher.User.CreatedOn,
		}
	}

	// Add business details if loaded
	if teacher.Business.ID != 0 {
		response.Business = &models.BusinessResponse{
			ID:        teacher.Business.ID,
			Name:      teacher.Business.Name,
			Slug:      teacher.Business.Slug,
			UserID:    teacher.Business.UserID,
			OwnerName: teacher.Business.OwnerName,
			Email:     teacher.Business.Email,
			Phone:     teacher.Business.Phone,
			Location:  teacher.Business.Location,
			Status:    teacher.Business.Status,
			CreatedOn: teacher.Business.CreatedOn,
		}
	}

	return response
}
