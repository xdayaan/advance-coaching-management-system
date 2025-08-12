package services

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/utils"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(req models.CreateUserRequest) (*models.UserResponse, string, error)
	Login(req models.LoginRequest) (*models.UserResponse, string, error)
	GetUsers(filters repository.UserFilters) ([]models.UserResponse, int64, error)
	GetUserByID(id uint) (*models.UserResponse, error)
	UpdateUser(id uint, updates map[string]interface{}) (*models.UserResponse, error)
	DeleteUser(id uint) error
	ValidateRole(role string) bool
	GetUsersByRole(role models.UserRole) ([]models.UserResponse, error)
	GetRoleStatistics() (map[models.UserRole]int64, error)
	PromoteUser(userID uint, newRole models.UserRole, promotedBy models.UserRole) error
	HasRolePermission(userRole models.UserRole, requiredRoles []models.UserRole) bool
	CanAccessRole(userRole models.UserRole, targetRole models.UserRole) bool
	ChangeUserStatus(userID uint, status int) error
	EmailExists(email string, excludeUserID ...uint) (bool, error)
	GetUserStats() (map[string]interface{}, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Register(req models.CreateUserRequest) (*models.UserResponse, string, error) {
	// Check if user already exists
	exists, err := s.repo.EmailExists(req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("error checking email existence: %w", err)
	}
	if exists {
		return nil, "", errors.New("email already exists")
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = models.RoleStudent
	}

	// Validate role
	if !req.Role.IsValid() {
		return nil, "", errors.New("invalid role provided")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("error hashing password: %w", err)
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Role:     req.Role,
		Status:   1, // Active by default
	}

	if err := s.repo.Create(user); err != nil {
		return nil, "", fmt.Errorf("error creating user: %w", err)
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", fmt.Errorf("error generating token: %w", err)
	}

	userResponse := s.toUserResponse(*user)
	return &userResponse, token, nil
}

func (s *userService) Login(req models.LoginRequest) (*models.UserResponse, string, error) {
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Check if user is active
	if user.Status != 1 {
		return nil, "", errors.New("account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", fmt.Errorf("error generating token: %w", err)
	}

	userResponse := s.toUserResponse(*user)
	return &userResponse, token, nil
}

func (s *userService) GetUsers(filters repository.UserFilters) ([]models.UserResponse, int64, error) {
	// Set default pagination if not provided
	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	users, total, err := s.repo.GetAll(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching users: %w", err)
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, s.toUserResponse(user))
	}

	return userResponses, total, nil
}

func (s *userService) GetUserByID(id uint) (*models.UserResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	userResponse := s.toUserResponse(*user)
	return &userResponse, nil
}

func (s *userService) UpdateUser(id uint, updates map[string]interface{}) (*models.UserResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Track if any updates were made
	hasUpdates := false

	// Apply updates with validation
	if name, ok := updates["name"].(string); ok && name != "" {
		user.Name = name
		hasUpdates = true
	}

	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
		hasUpdates = true
	}

	if email, ok := updates["email"].(string); ok && email != "" {
		// Check if email already exists for another user
		exists, err := s.repo.EmailExists(email, user.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking email existence: %w", err)
		}
		if exists {
			return nil, errors.New("email already exists")
		}
		user.Email = email
		hasUpdates = true
	}

	if roleStr, ok := updates["role"].(string); ok && roleStr != "" {
		role := models.UserRole(roleStr)
		if !role.IsValid() {
			return nil, errors.New("invalid role provided")
		}
		user.Role = role
		hasUpdates = true
	}

	if status, ok := updates["status"].(int); ok {
		if status < 0 || status > 1 {
			return nil, errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
		}
		user.Status = status
		hasUpdates = true
	}

	// Handle float64 status (from JSON)
	if status, ok := updates["status"].(float64); ok {
		statusInt := int(status)
		if statusInt < 0 || statusInt > 1 {
			return nil, errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
		}
		user.Status = statusInt
		hasUpdates = true
	}

	// Hash new password if provided
	if password, ok := updates["password"].(string); ok && password != "" {
		if len(password) < 6 {
			return nil, errors.New("password must be at least 6 characters")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("error hashing password: %w", err)
		}
		user.Password = string(hashedPassword)
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, errors.New("no valid updates provided")
	}

	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	userResponse := s.toUserResponse(*user)
	return &userResponse, nil
}

func (s *userService) DeleteUser(id uint) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

func (s *userService) ValidateRole(role string) bool {
	userRole := models.UserRole(role)
	return userRole.IsValid()
}

func (s *userService) GetUsersByRole(role models.UserRole) ([]models.UserResponse, error) {
	if !role.IsValid() {
		return nil, errors.New("invalid role provided")
	}

	users, err := s.repo.GetByRole(role)
	if err != nil {
		return nil, fmt.Errorf("error fetching users by role: %w", err)
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, s.toUserResponse(user))
	}

	return userResponses, nil
}

func (s *userService) GetRoleStatistics() (map[models.UserRole]int64, error) {
	stats := make(map[models.UserRole]int64)

	roles := []models.UserRole{
		models.RoleAdmin,
		models.RoleBusiness,
		models.RoleTeacher,
		models.RoleStudent,
	}

	for _, role := range roles {
		filters := repository.UserFilters{
			Role: string(role),
		}
		_, count, err := s.repo.GetAll(filters)
		if err != nil {
			return nil, fmt.Errorf("error getting statistics for role %s: %w", role, err)
		}
		stats[role] = count
	}

	return stats, nil
}

func (s *userService) PromoteUser(userID uint, newRole models.UserRole, promotedBy models.UserRole) error {
	if userID == 0 {
		return errors.New("invalid user ID")
	}

	if !newRole.IsValid() {
		return errors.New("invalid role provided")
	}

	// Check if the promoter has permission to promote to this role
	if !s.CanAccessRole(promotedBy, newRole) {
		return errors.New("insufficient permissions to promote to this role")
	}

	user, err := s.repo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Prevent demotion through this method
	if !s.CanAccessRole(newRole, user.Role) {
		return errors.New("cannot demote user through promotion method")
	}

	// Don't update if the role is the same
	if user.Role == newRole {
		return errors.New("user already has this role")
	}

	if err := s.repo.UpdateUserRole(userID, newRole); err != nil {
		return fmt.Errorf("error promoting user: %w", err)
	}

	return nil
}

// Helper method to check if a user has permission to access a role-based resource
func (s *userService) HasRolePermission(userRole models.UserRole, requiredRoles []models.UserRole) bool {
	for _, role := range requiredRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// Helper method to check role hierarchy (admin > business > teacher > student)
func (s *userService) CanAccessRole(userRole models.UserRole, targetRole models.UserRole) bool {
	roleHierarchy := map[models.UserRole]int{
		models.RoleAdmin:    4,
		models.RoleBusiness: 3,
		models.RoleTeacher:  2,
		models.RoleStudent:  1,
	}

	userLevel, userExists := roleHierarchy[userRole]
	targetLevel, targetExists := roleHierarchy[targetRole]

	if !userExists || !targetExists {
		return false
	}

	return userLevel >= targetLevel
}

func (s *userService) ChangeUserStatus(userID uint, status int) error {
	if userID == 0 {
		return errors.New("invalid user ID")
	}

	if status < 0 || status > 1 {
		return errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
	}

	_, err := s.repo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.repo.UpdateUserStatus(userID, status); err != nil {
		return fmt.Errorf("error updating user status: %w", err)
	}

	return nil
}

func (s *userService) EmailExists(email string, excludeUserID ...uint) (bool, error) {
	if email == "" {
		return false, errors.New("email cannot be empty")
	}

	return s.repo.EmailExists(email, excludeUserID...)
}

func (s *userService) GetUserStats() (map[string]interface{}, error) {
	stats, err := s.repo.GetUserStats()
	if err != nil {
		return nil, fmt.Errorf("error getting user statistics: %w", err)
	}

	return stats, nil
}

// Helper methods for user management
func (s *userService) GetActiveUsers() ([]models.UserResponse, error) {
	users, err := s.repo.GetUsersByStatus(1)
	if err != nil {
		return nil, fmt.Errorf("error fetching active users: %w", err)
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, s.toUserResponse(user))
	}

	return userResponses, nil
}

func (s *userService) GetInactiveUsers() ([]models.UserResponse, error) {
	users, err := s.repo.GetUsersByStatus(0)
	if err != nil {
		return nil, fmt.Errorf("error fetching inactive users: %w", err)
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, s.toUserResponse(user))
	}

	return userResponses, nil
}

func (s *userService) BulkUpdateUserStatus(userIDs []uint, status int) error {
	if len(userIDs) == 0 {
		return errors.New("no user IDs provided")
	}

	if status < 0 || status > 1 {
		return errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
	}

	for _, userID := range userIDs {
		if err := s.ChangeUserStatus(userID, status); err != nil {
			return fmt.Errorf("error updating status for user ID %d: %w", userID, err)
		}
	}

	return nil
}

// Helper method to check if a user can perform actions on another user
func (s *userService) CanManageUser(managerRole models.UserRole, targetUserID uint) (bool, error) {
	targetUser, err := s.repo.GetByID(targetUserID)
	if err != nil {
		return false, errors.New("target user not found")
	}

	return s.CanAccessRole(managerRole, targetUser.Role), nil
}

func (s *userService) toUserResponse(user models.User) models.UserResponse {
	return models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		Status:    user.Status,
		CreatedOn: user.CreatedOn,
	}
}

// Validation helpers
func (s *userService) ValidateUserRole(userRole models.UserRole, requiredRoles ...models.UserRole) bool {
	for _, role := range requiredRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

func (s *userService) GetRoleHierarchy() map[models.UserRole]int {
	return map[models.UserRole]int{
		models.RoleAdmin:    4,
		models.RoleBusiness: 3,
		models.RoleTeacher:  2,
		models.RoleStudent:  1,
	}
}

func (s *userService) GetAllRoles() []models.UserRole {
	return []models.UserRole{
		models.RoleAdmin,
		models.RoleBusiness,
		models.RoleTeacher,
		models.RoleStudent,
	}
}
