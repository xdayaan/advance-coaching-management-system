package repository

import (
	"backend/internal/models"
	"backend/pkg/database"
	"fmt"

	"gorm.io/gorm"
)

func (r *userRepository) UpdateUser(user *models.User) error {
	return r.Update(user)
}

type UserRepository interface {
	// Basic CRUD operations
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	GetAll(filters UserFilters) ([]models.User, int64, error)
	Update(user *models.User) error
	UpdateUser(user *models.User) error
	Delete(id uint) error

	// Role-based operations
	GetByRole(role models.UserRole) ([]models.User, error)
	UpdateUserRole(userID uint, newRole models.UserRole) error

	// Status operations
	GetActiveByEmail(email string) (*models.User, error)
	GetUsersByStatus(status int) ([]models.User, error)
	UpdateUserStatus(userID uint, status int) error

	// Statistics and reporting
	GetUserStats() (map[string]interface{}, error)
	GetRoleCount(role models.UserRole) (int64, error)
	GetStatusCount(status int) (int64, error)

	// Validation and utility
	EmailExists(email string, excludeUserID ...uint) (bool, error)
	GetUsersCount() (int64, error)

	// Bulk operations
	BulkUpdateStatus(userIDs []uint, status int) error
	BulkDelete(userIDs []uint) error

	// Transactional operations
	CreateUserInTransaction(tx *gorm.DB, user *models.User) error

	// Advanced queries
	SearchUsers(searchTerm string, limit int) ([]models.User, error)
	GetRecentUsers(limit int) ([]models.User, error)
	GetUsersByDateRange(startDate, endDate string) ([]models.User, error)
}

type UserFilters struct {
	Role      string `form:"role" json:"role"`
	Status    *int   `form:"status" json:"status"`
	Search    string `form:"search" json:"search"`
	Page      int    `form:"page" json:"page"`
	Limit     int    `form:"limit" json:"limit"`
	SortBy    string `form:"sort_by" json:"sort_by"`       // created_on, name, email
	SortOrder string `form:"sort_order" json:"sort_order"` // asc, desc
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.DB,
	}
}

// Basic CRUD operations

func (r *userRepository) Create(user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	return r.db.Create(user).Error
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetActiveByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var user models.User
	err := r.db.Where("email = ? AND status = 1", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(filters UserFilters) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if filters.Role != "" {
		role := models.UserRole(filters.Role)
		if role.IsValid() {
			query = query.Where("role = ?", filters.Role)
		}
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?",
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
			"created_on": true,
			"updated_on": true,
			"name":       true,
			"email":      true,
			"role":       true,
			"status":     true,
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

	err := query.Find(&users).Error
	return users, total, err
}

func (r *userRepository) Update(user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	if user.ID == 0 {
		return fmt.Errorf("user ID cannot be zero")
	}
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid user ID")
	}
	return r.db.Delete(&models.User{}, id).Error
}

// Role-based operations

func (r *userRepository) GetByRole(role models.UserRole) ([]models.User, error) {
	var users []models.User

	if !role.IsValid() {
		return users, gorm.ErrInvalidValue
	}

	err := r.db.Where("role = ?", string(role)).Order("created_on DESC").Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateUserRole(userID uint, newRole models.UserRole) error {
	if userID == 0 {
		return fmt.Errorf("invalid user ID")
	}
	if !newRole.IsValid() {
		return gorm.ErrInvalidValue
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role", newRole).Error
}

func (r *userRepository) GetRoleCount(role models.UserRole) (int64, error) {
	if !role.IsValid() {
		return 0, gorm.ErrInvalidValue
	}

	var count int64
	err := r.db.Model(&models.User{}).Where("role = ?", string(role)).Count(&count).Error
	return count, err
}

// Status operations

func (r *userRepository) GetUsersByStatus(status int) ([]models.User, error) {
	var users []models.User

	if status < 0 || status > 1 {
		return users, gorm.ErrInvalidValue
	}

	err := r.db.Where("status = ?", status).Order("created_on DESC").Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateUserStatus(userID uint, status int) error {
	if userID == 0 {
		return fmt.Errorf("invalid user ID")
	}
	if status < 0 || status > 1 {
		return gorm.ErrInvalidValue
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error
}

func (r *userRepository) GetStatusCount(status int) (int64, error) {
	if status < 0 || status > 1 {
		return 0, gorm.ErrInvalidValue
	}

	var count int64
	err := r.db.Model(&models.User{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// Statistics and reporting

func (r *userRepository) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users
	var totalUsers int64
	if err := r.db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// Active users
	var activeUsers int64
	if err := r.db.Model(&models.User{}).Where("status = 1").Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	stats["active_users"] = activeUsers

	// Inactive users
	var inactiveUsers int64
	if err := r.db.Model(&models.User{}).Where("status = 0").Count(&inactiveUsers).Error; err != nil {
		return nil, err
	}
	stats["inactive_users"] = inactiveUsers

	// Users by role
	roleStats := make(map[string]int64)
	roles := []models.UserRole{models.RoleAdmin, models.RoleBusiness, models.RoleTeacher, models.RoleStudent}

	for _, role := range roles {
		var count int64
		if err := r.db.Model(&models.User{}).Where("role = ?", string(role)).Count(&count).Error; err != nil {
			return nil, err
		}
		roleStats[string(role)] = count
	}
	stats["by_role"] = roleStats

	// Active users by role
	activeRoleStats := make(map[string]int64)
	for _, role := range roles {
		var count int64
		if err := r.db.Model(&models.User{}).Where("role = ? AND status = 1", string(role)).Count(&count).Error; err != nil {
			return nil, err
		}
		activeRoleStats[string(role)] = count
	}
	stats["active_by_role"] = activeRoleStats

	return stats, nil
}

func (r *userRepository) GetUsersCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// Validation and utility

func (r *userRepository) EmailExists(email string, excludeUserID ...uint) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("email cannot be empty")
	}

	var count int64
	query := r.db.Model(&models.User{}).Where("email = ?", email)

	if len(excludeUserID) > 0 && excludeUserID[0] > 0 {
		query = query.Where("id != ?", excludeUserID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// Bulk operations

func (r *userRepository) BulkUpdateStatus(userIDs []uint, status int) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("no user IDs provided")
	}
	if status < 0 || status > 1 {
		return gorm.ErrInvalidValue
	}

	return r.db.Model(&models.User{}).Where("id IN ?", userIDs).Update("status", status).Error
}

func (r *userRepository) BulkDelete(userIDs []uint) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("no user IDs provided")
	}

	return r.db.Where("id IN ?", userIDs).Delete(&models.User{}).Error
}

// Advanced queries

func (r *userRepository) SearchUsers(searchTerm string, limit int) ([]models.User, error) {
	if searchTerm == "" {
		return []models.User{}, nil
	}

	var users []models.User
	query := r.db.Where("name ILIKE ? OR email ILIKE ?", "%"+searchTerm+"%", "%"+searchTerm+"%").
		Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&users).Error
	return users, err
}

func (r *userRepository) GetRecentUsers(limit int) ([]models.User, error) {
	var users []models.User
	query := r.db.Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(10) // default limit
	}

	err := query.Find(&users).Error
	return users, err
}

func (r *userRepository) GetUsersByDateRange(startDate, endDate string) ([]models.User, error) {
	if startDate == "" || endDate == "" {
		return nil, fmt.Errorf("start date and end date cannot be empty")
	}

	var users []models.User
	err := r.db.Where("created_on BETWEEN ? AND ?", startDate, endDate).
		Order("created_on DESC").Find(&users).Error
	return users, err
}

// Additional utility methods

// GetUsersByRoleAndStatus gets users by both role and status
func (r *userRepository) GetUsersByRoleAndStatus(role models.UserRole, status int) ([]models.User, error) {
	if !role.IsValid() {
		return nil, gorm.ErrInvalidValue
	}
	if status < 0 || status > 1 {
		return nil, gorm.ErrInvalidValue
	}

	var users []models.User
	err := r.db.Where("role = ? AND status = ?", string(role), status).
		Order("created_on DESC").Find(&users).Error
	return users, err
}

// UpdateUserEmail updates only the user's email
func (r *userRepository) UpdateUserEmail(userID uint, newEmail string) error {
	if userID == 0 {
		return fmt.Errorf("invalid user ID")
	}
	if newEmail == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Check if email already exists for another user
	exists, err := r.EmailExists(newEmail, userID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email already exists")
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("email", newEmail).Error
}

// UpdateUserPassword updates only the user's password
func (r *userRepository) UpdateUserPassword(userID uint, hashedPassword string) error {
	if userID == 0 {
		return fmt.Errorf("invalid user ID")
	}
	if hashedPassword == "" {
		return fmt.Errorf("password cannot be empty")
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// GetUserByEmailAndStatus gets user by email and specific status
func (r *userRepository) GetUserByEmailAndStatus(email string, status int) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if status < 0 || status > 1 {
		return nil, gorm.ErrInvalidValue
	}

	var user models.User
	err := r.db.Where("email = ? AND status = ?", email, status).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsersWithPagination gets users with simple pagination (offset/limit)
func (r *userRepository) GetUsersWithPagination(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Get total count
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	err := r.db.Offset(offset).Limit(limit).Order("created_on DESC").Find(&users).Error
	return users, total, err
}

// SoftDeleteUser performs soft delete (sets status to inactive instead of deleting)
func (r *userRepository) SoftDeleteUser(userID uint) error {
	if userID == 0 {
		return fmt.Errorf("invalid user ID")
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("status", 0).Error
}

// RestoreUser restores a soft-deleted user (sets status back to active)
func (r *userRepository) RestoreUser(userID uint) error {
	if userID == 0 {
		return fmt.Errorf("invalid user ID")
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("status", 1).Error
}

// GetUserCountByDateRange gets count of users created within a date range
func (r *userRepository) GetUserCountByDateRange(startDate, endDate string) (int64, error) {
	if startDate == "" || endDate == "" {
		return 0, fmt.Errorf("start date and end date cannot be empty")
	}

	var count int64
	err := r.db.Model(&models.User{}).Where("created_on BETWEEN ? AND ?", startDate, endDate).Count(&count).Error
	return count, err
}

// GetTopUsersByRole gets top N users by role (most recently created)
func (r *userRepository) GetTopUsersByRole(role models.UserRole, limit int) ([]models.User, error) {
	if !role.IsValid() {
		return nil, gorm.ErrInvalidValue
	}

	var users []models.User
	query := r.db.Where("role = ?", string(role)).Order("created_on DESC")

	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(10) // default limit
	}

	err := query.Find(&users).Error
	return users, err
}

// BatchUpdateUsers updates multiple users with the same data
func (r *userRepository) BatchUpdateUsers(userIDs []uint, updates map[string]interface{}) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("no user IDs provided")
	}
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	return r.db.Model(&models.User{}).Where("id IN ?", userIDs).Updates(updates).Error
}

// GetUsersByMultipleRoles gets users by multiple roles
func (r *userRepository) GetUsersByMultipleRoles(roles []models.UserRole) ([]models.User, error) {
	if len(roles) == 0 {
		return []models.User{}, nil
	}

	// Convert roles to strings
	var roleStrings []string
	for _, role := range roles {
		if role.IsValid() {
			roleStrings = append(roleStrings, string(role))
		}
	}

	if len(roleStrings) == 0 {
		return []models.User{}, nil
	}

	var users []models.User
	err := r.db.Where("role IN ?", roleStrings).Order("created_on DESC").Find(&users).Error
	return users, err
}

// GetUsersByMultipleStatuses gets users by multiple statuses
func (r *userRepository) GetUsersByMultipleStatuses(statuses []int) ([]models.User, error) {
	if len(statuses) == 0 {
		return []models.User{}, nil
	}

	// Validate statuses
	var validStatuses []int
	for _, status := range statuses {
		if status >= 0 && status <= 1 {
			validStatuses = append(validStatuses, status)
		}
	}

	if len(validStatuses) == 0 {
		return []models.User{}, nil
	}

	var users []models.User
	err := r.db.Where("status IN ?", validStatuses).Order("created_on DESC").Find(&users).Error
	return users, err
}

// CheckUserExists checks if a user exists by ID
func (r *userRepository) CheckUserExists(userID uint) (bool, error) {
	if userID == 0 {
		return false, fmt.Errorf("invalid user ID")
	}

	var count int64
	err := r.db.Model(&models.User{}).Where("id = ?", userID).Count(&count).Error
	return count > 0, err
}

// GetUserIDsByRole gets only user IDs for a specific role (lightweight query)
func (r *userRepository) GetUserIDsByRole(role models.UserRole) ([]uint, error) {
	if !role.IsValid() {
		return nil, gorm.ErrInvalidValue
	}

	var userIDs []uint
	err := r.db.Model(&models.User{}).Where("role = ?", string(role)).Pluck("id", &userIDs).Error
	return userIDs, err
}

// GetUserIDsByStatus gets only user IDs for a specific status (lightweight query)
func (r *userRepository) GetUserIDsByStatus(status int) ([]uint, error) {
	if status < 0 || status > 1 {
		return nil, gorm.ErrInvalidValue
	}

	var userIDs []uint
	err := r.db.Model(&models.User{}).Where("status = ?", status).Pluck("id", &userIDs).Error
	return userIDs, err
}

// GetUserEmailsByRole gets only user emails for a specific role (for notifications, etc.)
func (r *userRepository) GetUserEmailsByRole(role models.UserRole) ([]string, error) {
	if !role.IsValid() {
		return nil, gorm.ErrInvalidValue
	}

	var emails []string
	err := r.db.Model(&models.User{}).Where("role = ? AND status = 1", string(role)).Pluck("email", &emails).Error
	return emails, err
}

// Transaction support methods

// CreateUserInTransaction creates a user within a transaction
func (r *userRepository) CreateUserInTransaction(tx *gorm.DB, user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	return tx.Create(user).Error
}

// UpdateUserInTransaction updates a user within a transaction
func (r *userRepository) UpdateUserInTransaction(tx *gorm.DB, user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	if user.ID == 0 {
		return fmt.Errorf("user ID cannot be zero")
	}
	return tx.Save(user).Error
}

// BeginTransaction starts a new database transaction
func (r *userRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Performance and analytics methods

// GetDailyUserRegistrations gets count of users registered per day for the last N days
func (r *userRepository) GetDailyUserRegistrations(days int) (map[string]int64, error) {
	if days <= 0 {
		days = 7 // default to last 7 days
	}

	rows, err := r.db.Raw(`
		SELECT DATE(created_on) as date, COUNT(*) as count 
		FROM users 
		WHERE created_on >= NOW() - INTERVAL ? DAY 
		GROUP BY DATE(created_on) 
		ORDER BY date DESC
	`, days).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var date string
		var count int64
		if err := rows.Scan(&date, &count); err != nil {
			continue
		}
		result[date] = count
	}

	return result, nil
}

// GetMonthlyUserStats gets user statistics grouped by month
func (r *userRepository) GetMonthlyUserStats(months int) (map[string]map[string]int64, error) {
	if months <= 0 {
		months = 6 // default to last 6 months
	}

	rows, err := r.db.Raw(`
		SELECT 
			DATE_TRUNC('month', created_on) as month,
			role,
			COUNT(*) as count 
		FROM users 
		WHERE created_on >= NOW() - INTERVAL ? MONTH 
		GROUP BY DATE_TRUNC('month', created_on), role 
		ORDER BY month DESC, role
	`, months).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]int64)
	for rows.Next() {
		var month, role string
		var count int64
		if err := rows.Scan(&month, &role, &count); err != nil {
			continue
		}

		if result[month] == nil {
			result[month] = make(map[string]int64)
		}
		result[month][role] = count
	}

	return result, nil
}
