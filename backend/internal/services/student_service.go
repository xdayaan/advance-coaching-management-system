package services

import (
	"backend/internal/models"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type StudentService interface {
	// Basic CRUD operations
	CreateStudent(req models.CreateStudentRequest) (*models.StudentResponse, error)
	GetStudentByID(id uint) (*models.StudentResponse, error)
	GetStudentByUserID(userID uint) (*models.StudentResponse, error)
	GetStudents(filters repository.StudentFilters) ([]models.StudentResponse, int64, error)
	UpdateStudent(studentID uint, updates map[string]interface{}) (*models.StudentResponse, error)
	DeleteStudent(studentID uint) error

	// Business specific operations
	GetStudentsByBusiness(businessID uint, filters repository.StudentFilters) ([]models.StudentResponse, int64, error)
	GetActiveStudentsByBusiness(businessID uint) ([]models.StudentResponse, error)
	GetInactiveStudentsByBusiness(businessID uint) ([]models.StudentResponse, error)

	// Status operations
	ChangeStudentStatus(studentID uint, status int) error
	GetActiveStudents() ([]models.StudentResponse, error)
	GetInactiveStudents() ([]models.StudentResponse, error)

	// Search
	SearchStudents(searchTerm string, limit int, businessID ...uint) ([]models.StudentResponse, error)
	SearchStudentsByBusiness(businessID uint, searchTerm string, limit int) ([]models.StudentResponse, error)

	// Statistics
	GetStudentStats(businessID ...uint) (map[string]interface{}, error)
	GetGuardianStats(businessID ...uint) (map[string]interface{}, error)

	// Bulk operations
	BulkUpdateStudentStatus(studentIDs []uint, status int) error

	// Validation
	ValidateCreateStudentRequest(req models.CreateStudentRequest) error
	ValidateUpdateStudentRequest(req models.UpdateStudentRequest) error
}

type studentService struct {
	studentRepo  repository.StudentRepository
	userRepo     repository.UserRepository
	businessRepo repository.BusinessRepository
}

func NewStudentService(studentRepo repository.StudentRepository, userRepo repository.UserRepository, businessRepo repository.BusinessRepository) StudentService {
	return &studentService{
		studentRepo:  studentRepo,
		userRepo:     userRepo,
		businessRepo: businessRepo,
	}
}

func (s *studentService) CreateStudent(req models.CreateStudentRequest) (*models.StudentResponse, error) {
	// Validate request
	if err := s.ValidateCreateStudentRequest(req); err != nil {
		return nil, err
	}

	// Check if user exists and is not already a student
	if _, err := s.userRepo.GetByID(req.UserID); err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is already a student
	exists, err := s.studentRepo.StudentUserExists(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user is already a student")
	}
	if exists {
		return nil, fmt.Errorf("user is already a student")
	}

	// Check if business exists
	_, err = s.businessRepo.GetByID(req.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("business not found")
	}

	// Initialize information if nil
	if req.Information == nil {
		req.Information = make(models.JSONB)
	}

	// Create student
	student := &models.Student{
		Name:           req.Name,
		UserID:         req.UserID,
		BusinessID:     req.BusinessID,
		GuardianName:   req.GuardianName,
		GuardianNumber: req.GuardianNumber,
		GuardianEmail:  req.GuardianEmail,
		Information:    req.Information,
		Status:         1, // Active by default
	}

	if err := s.studentRepo.Create(student); err != nil {
		return nil, fmt.Errorf("failed to create student: %v", err)
	}

	// Get student with relations
	studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created student")
	}

	return s.toStudentResponse(studentWithRelations), nil
}

func (s *studentService) GetStudentByID(id uint) (*models.StudentResponse, error) {
	student, err := s.studentRepo.GetStudentWithRelations(id)
	if err != nil {
		return nil, fmt.Errorf("student not found")
	}

	return s.toStudentResponse(student), nil
}

func (s *studentService) GetStudentByUserID(userID uint) (*models.StudentResponse, error) {
	student, err := s.studentRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("student profile not found")
	}

	// Get with relations
	studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student details")
	}

	return s.toStudentResponse(studentWithRelations), nil
}

func (s *studentService) GetStudents(filters repository.StudentFilters) ([]models.StudentResponse, int64, error) {
	// Set default pagination
	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Limit == 0 {
		filters.Limit = 10
	}

	students, total, err := s.studentRepo.GetAllWithRelations(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get students: %v", err)
	}

	var responses []models.StudentResponse
	for _, student := range students {
		responses = append(responses, *s.toStudentResponse(&student))
	}

	return responses, total, nil
}

func (s *studentService) UpdateStudent(studentID uint, updates map[string]interface{}) (*models.StudentResponse, error) {
	// Get existing student
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("student not found")
	}

	// Update fields
	if name, ok := updates["name"]; ok {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			student.Name = nameStr
		}
	}

	if guardianName, ok := updates["guardian_name"]; ok {
		if guardianNameStr, ok := guardianName.(string); ok {
			student.GuardianName = guardianNameStr
		}
	}

	if guardianNumber, ok := updates["guardian_number"]; ok {
		if guardianNumberStr, ok := guardianNumber.(string); ok {
			student.GuardianNumber = guardianNumberStr
		}
	}

	if guardianEmail, ok := updates["guardian_email"]; ok {
		if guardianEmailStr, ok := guardianEmail.(string); ok {
			student.GuardianEmail = guardianEmailStr
		}
	}

	if information, ok := updates["information"]; ok {
		if infoMap, ok := information.(models.JSONB); ok {
			student.Information = infoMap
		}
	}

	if status, ok := updates["status"]; ok {
		if statusInt, ok := status.(int); ok && (statusInt == 0 || statusInt == 1) {
			student.Status = statusInt
		}
	}

	// Save updates
	if err := s.studentRepo.Update(student); err != nil {
		return nil, fmt.Errorf("failed to update student: %v", err)
	}

	// Get updated student with relations
	updatedStudent, err := s.studentRepo.GetStudentWithRelations(student.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated student")
	}

	return s.toStudentResponse(updatedStudent), nil
}

func (s *studentService) DeleteStudent(studentID uint) error {
	// Check if student exists
	_, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return fmt.Errorf("student not found")
	}

	if err := s.studentRepo.Delete(studentID); err != nil {
		return fmt.Errorf("failed to delete student: %v", err)
	}

	return nil
}

func (s *studentService) GetStudentsByBusiness(businessID uint, filters repository.StudentFilters) ([]models.StudentResponse, int64, error) {
	// Set default pagination
	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Limit == 0 {
		filters.Limit = 10
	}

	students, total, err := s.studentRepo.GetByBusinessID(businessID, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get students by business: %v", err)
	}

	var responses []models.StudentResponse
	for _, student := range students {
		// Get with relations for complete data
		studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
		if err != nil {
			continue // Skip this one if there's an issue
		}
		responses = append(responses, *s.toStudentResponse(studentWithRelations))
	}

	return responses, total, nil
}

func (s *studentService) GetActiveStudentsByBusiness(businessID uint) ([]models.StudentResponse, error) {
	students, err := s.studentRepo.GetActiveStudentsByBusiness(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active students: %v", err)
	}

	var responses []models.StudentResponse
	for _, student := range students {
		studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toStudentResponse(studentWithRelations))
	}

	return responses, nil
}

func (s *studentService) GetInactiveStudentsByBusiness(businessID uint) ([]models.StudentResponse, error) {
	students, err := s.studentRepo.GetInactiveStudentsByBusiness(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive students: %v", err)
	}

	var responses []models.StudentResponse
	for _, student := range students {
		studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toStudentResponse(studentWithRelations))
	}

	return responses, nil
}

func (s *studentService) ChangeStudentStatus(studentID uint, status int) error {
	// Check if student exists
	_, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return fmt.Errorf("student not found")
	}

	if err := s.studentRepo.UpdateStudentStatus(studentID, status); err != nil {
		return fmt.Errorf("failed to update student status: %v", err)
	}

	return nil
}

func (s *studentService) GetActiveStudents() ([]models.StudentResponse, error) {
	students, err := s.studentRepo.GetActiveStudents()
	if err != nil {
		return nil, fmt.Errorf("failed to get active students: %v", err)
	}

	var responses []models.StudentResponse
	for _, student := range students {
		studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toStudentResponse(studentWithRelations))
	}

	return responses, nil
}

func (s *studentService) GetInactiveStudents() ([]models.StudentResponse, error) {
	students, err := s.studentRepo.GetInactiveStudents()
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive students: %v", err)
	}

	var responses []models.StudentResponse
	for _, student := range students {
		studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toStudentResponse(studentWithRelations))
	}

	return responses, nil
}

func (s *studentService) SearchStudents(searchTerm string, limit int, businessID ...uint) ([]models.StudentResponse, error) {
	students, err := s.studentRepo.SearchStudents(searchTerm, limit, businessID...)
	if err != nil {
		return nil, fmt.Errorf("failed to search students: %v", err)
	}

	var responses []models.StudentResponse
	for _, student := range students {
		studentWithRelations, err := s.studentRepo.GetStudentWithRelations(student.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toStudentResponse(studentWithRelations))
	}

	return responses, nil
}

func (s *studentService) SearchStudentsByBusiness(businessID uint, searchTerm string, limit int) ([]models.StudentResponse, error) {
	return s.SearchStudents(searchTerm, limit, businessID)
}

func (s *studentService) GetStudentStats(businessID ...uint) (map[string]interface{}, error) {
	return s.studentRepo.GetStudentStats(businessID...)
}

func (s *studentService) GetGuardianStats(businessID ...uint) (map[string]interface{}, error) {
	return s.studentRepo.GetGuardianStats(businessID...)
}

func (s *studentService) BulkUpdateStudentStatus(studentIDs []uint, status int) error {
	if len(studentIDs) == 0 {
		return fmt.Errorf("no student IDs provided")
	}

	if err := s.studentRepo.BulkUpdateStatus(studentIDs, status); err != nil {
		return fmt.Errorf("failed to bulk update student status: %v", err)
	}

	return nil
}

func (s *studentService) ValidateCreateStudentRequest(req models.CreateStudentRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if req.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}

	if req.BusinessID == 0 {
		return fmt.Errorf("business ID is required")
	}

	// Validate guardian email if provided
	if req.GuardianEmail != "" {
		// Basic email validation
		if !strings.Contains(req.GuardianEmail, "@") {
			return fmt.Errorf("invalid guardian email format")
		}
	}

	return nil
}

func (s *studentService) ValidateUpdateStudentRequest(req models.UpdateStudentRequest) error {
	if req.Name != "" && strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if req.Status != nil && (*req.Status < 0 || *req.Status > 1) {
		return fmt.Errorf("invalid status value")
	}

	// Validate guardian email if provided
	if req.GuardianEmail != "" && !strings.Contains(req.GuardianEmail, "@") {
		return fmt.Errorf("invalid guardian email format")
	}

	return nil
}

// Helper methods
func (s *studentService) toStudentResponse(student *models.Student) *models.StudentResponse {
	response := &models.StudentResponse{
		ID:             student.ID,
		Name:           student.Name,
		UserID:         student.UserID,
		BusinessID:     student.BusinessID,
		GuardianName:   student.GuardianName,
		GuardianNumber: student.GuardianNumber,
		GuardianEmail:  student.GuardianEmail,
		Information:    student.Information,
		Status:         student.Status,
		CreatedOn:      student.CreatedOn,
		UpdatedOn:      student.UpdatedOn,
	}

	// Add user details if loaded
	if student.User.ID != 0 {
		response.User = &models.UserResponse{
			ID:        student.User.ID,
			Name:      student.User.Name,
			Email:     student.User.Email,
			Role:      student.User.Role,
			Status:    student.User.Status,
			CreatedOn: student.User.CreatedOn,
		}
	}

	// Add business details if loaded
	if student.Business.ID != 0 {
		response.Business = &models.BusinessResponse{
			ID:        student.Business.ID,
			Name:      student.Business.Name,
			Slug:      student.Business.Slug,
			UserID:    student.Business.UserID,
			OwnerName: student.Business.OwnerName,
			Email:     student.Business.Email,
			Phone:     student.Business.Phone,
			Location:  student.Business.Location,
			Status:    student.Business.Status,
			CreatedOn: student.Business.CreatedOn,
		}
	}

	return response
}
