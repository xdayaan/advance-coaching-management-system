package services

import (
	"backend/internal/models"
	"backend/internal/repository"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func generateSlugFromName(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

type BusinessService interface {
	CreateBusiness(req models.CreateBusinessRequest) (*models.BusinessResponse, error)
	GetBusinesses(filters repository.BusinessFilters) ([]models.BusinessResponse, int64, error)
	GetBusinessByID(id uint) (*models.BusinessResponse, error)
	GetBusinessBySlug(slug string) (*models.BusinessResponse, error)
	GetBusinessByUserID(userID uint) (*models.BusinessResponse, error)
	UpdateBusiness(id uint, updates map[string]interface{}) (*models.BusinessResponse, error)
	DeleteBusiness(id uint) error
	GetActiveBusinesses() ([]models.BusinessResponse, error)
	GetInactiveBusinesses() ([]models.BusinessResponse, error)
	ChangeBusinessStatus(businessID uint, status int) error
	AssignPackage(businessID, packageID uint) error
	RemovePackage(businessID uint) error
	GetBusinessesByPackage(packageID uint) ([]models.BusinessResponse, error)
	GetBusinessesWithoutPackage() ([]models.BusinessResponse, error)
	BusinessEmailExists(email string, excludeBusinessID ...uint) (bool, error)
	BusinessNameExists(name string, excludeBusinessID ...uint) (bool, error)
	GetBusinessStats() (map[string]interface{}, error)
	GetLocationStats() (map[string]int64, error)
	GetPackageDistribution() (map[string]int64, error)
	SearchBusinesses(searchTerm string, limit int) ([]models.BusinessResponse, error)
	BulkUpdateBusinessStatus(businessIDs []uint, status int) error
	BulkAssignPackage(businessIDs []uint, packageID uint) error
	GetBusinessesByLocation(location string) ([]models.BusinessResponse, error)
	GetBusinessLocations() ([]string, error)
}

type businessService struct {
	businessRepo repository.BusinessRepository
	userRepo     repository.UserRepository
	packageRepo  repository.PackageRepository
}

func NewBusinessService(businessRepo repository.BusinessRepository, userRepo repository.UserRepository, packageRepo repository.PackageRepository) BusinessService {
	return &businessService{
		businessRepo: businessRepo,
		userRepo:     userRepo,
		packageRepo:  packageRepo,
	}
}

func (s *businessService) CreateBusiness(req models.CreateBusinessRequest) (*models.BusinessResponse, error) {
	// Check if business email already exists
	exists, err := s.businessRepo.BusinessEmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking business email existence: %w", err)
	}
	if exists {
		return nil, errors.New("business email already exists")
	}

	// Check if user email already exists (for creating user account)
	exists, err = s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking user email existence: %w", err)
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Validate package if provided
	if req.PackageID != nil {
		_, err := s.packageRepo.GetByID(*req.PackageID)
		if err != nil {
			return nil, errors.New("invalid package ID")
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Start transaction
	tx := s.businessRepo.BeginTransaction()

	// Create user account first
	user := &models.User{
		Name:     req.OwnerName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Role:     models.RoleBusiness,
		Status:   1, // Active by default
	}

	if err := s.userRepo.CreateUserInTransaction(tx, user); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating user account: %w", err)
	}

	slug := generateSlugFromName(req.Slug)

	fmt.Println("Slug:", slug)

	originalSlug := slug
	counter := 1
	for {
		exists, err := s.businessRepo.BusinessSlugExists(slug)
		if err != nil {
			return nil, fmt.Errorf("error checking slug existence: %w", err)
		}
		if !exists {
			break
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	// Create business
	business := &models.Business{
		Name:      req.Name,
		Slug:      slug, // Add this line
		UserID:    user.ID,
		OwnerName: req.OwnerName,
		PackageID: req.PackageID,
		Email:     req.Email,
		Phone:     req.Phone,
		Location:  req.Location,
		Password:  string(hashedPassword),
		Status:    1,
	}

	if err := s.businessRepo.CreateWithTransaction(tx, business); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating business: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	businessResponse := s.toBusinessResponse(*business)
	return &businessResponse, nil
}

func (s *businessService) GetBusinesses(filters repository.BusinessFilters) ([]models.BusinessResponse, int64, error) {
	// Set default pagination if not provided
	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	businesses, total, err := s.businessRepo.GetAllWithRelations(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching businesses: %w", err)
	}

	var businessResponses []models.BusinessResponse
	for _, business := range businesses {
		businessResponses = append(businessResponses, s.toBusinessResponseWithRelations(business))
	}

	return businessResponses, total, nil
}

func (s *businessService) GetBusinessByID(id uint) (*models.BusinessResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid business ID")
	}

	business, err := s.businessRepo.GetBusinessWithRelations(id)
	if err != nil {
		return nil, errors.New("business not found")
	}

	businessResponse := s.toBusinessResponseWithRelations(*business)
	return &businessResponse, nil
}

func (s *businessService) GetBusinessBySlug(slug string) (*models.BusinessResponse, error) {
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	business, err := s.businessRepo.GetBySlugWithRelations(slug)
	if err != nil {
		return nil, errors.New("business not found")
	}

	businessResponse := s.toBusinessResponseWithRelations(*business)
	return &businessResponse, nil
}

func (s *businessService) GetBusinessByUserID(userID uint) (*models.BusinessResponse, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}

	business, err := s.businessRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("business not found")
	}

	businessResponse := s.toBusinessResponse(*business)
	return &businessResponse, nil
}

func (s *businessService) UpdateBusiness(id uint, updates map[string]interface{}) (*models.BusinessResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid business ID")
	}

	business, err := s.businessRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("business not found")
	}

	// Track if any updates were made
	hasUpdates := false
	hasUserUpdates := false

	// Prepare user updates
	userUpdates := make(map[string]interface{})

	// Apply updates with validation
	if name, ok := updates["name"].(string); ok && name != "" {
		// Check if name already exists for another business
		exists, err := s.businessRepo.BusinessNameExists(name, business.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking business name existence: %w", err)
		}
		if exists {
			return nil, errors.New("business name already exists")
		}
		business.Name = name
		hasUpdates = true
	}

	if ownerName, ok := updates["owner_name"].(string); ok && ownerName != "" {
		business.OwnerName = ownerName
		userUpdates["name"] = ownerName
		hasUpdates = true
		hasUserUpdates = true
	}

	if email, ok := updates["email"].(string); ok && email != "" {
		// Check if email already exists for another business
		exists, err := s.businessRepo.BusinessEmailExists(email, business.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking business email existence: %w", err)
		}
		if exists {
			return nil, errors.New("business email already exists")
		}

		// Check if email already exists for another user
		exists, err = s.userRepo.EmailExists(email, business.UserID)
		if err != nil {
			return nil, fmt.Errorf("error checking user email existence: %w", err)
		}
		if exists {
			return nil, errors.New("user with this email already exists")
		}

		business.Email = email
		userUpdates["email"] = email
		hasUpdates = true
		hasUserUpdates = true
	}

	if phone, ok := updates["phone"].(string); ok {
		business.Phone = phone
		userUpdates["phone"] = phone
		hasUpdates = true
		hasUserUpdates = true
	}

	if location, ok := updates["location"].(string); ok {
		business.Location = location
		hasUpdates = true
	}

	if packageID, ok := updates["package_id"]; ok {
		if packageID == nil {
			business.PackageID = nil
		} else {
			packageIDUint := uint(packageID.(float64))
			// Validate package exists
			_, err := s.packageRepo.GetByID(packageIDUint)
			if err != nil {
				return nil, errors.New("invalid package ID")
			}
			business.PackageID = &packageIDUint
		}
		hasUpdates = true
	}

	if status, ok := updates["status"].(int); ok {
		if status < 0 || status > 1 {
			return nil, errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
		}
		business.Status = status
		userUpdates["status"] = status
		hasUpdates = true
		hasUserUpdates = true
	}

	// Handle float64 status (from JSON)
	if status, ok := updates["status"].(float64); ok {
		statusInt := int(status)
		if statusInt < 0 || statusInt > 1 {
			return nil, errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
		}
		business.Status = statusInt
		userUpdates["status"] = statusInt
		hasUpdates = true
		hasUserUpdates = true
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
		business.Password = string(hashedPassword)
		userUpdates["password"] = string(hashedPassword)
		hasUpdates = true
		hasUserUpdates = true
	}

	if !hasUpdates {
		return nil, errors.New("no valid updates provided")
	}

	// Start transaction
	tx := s.businessRepo.BeginTransaction()

	// Update business
	if err := s.businessRepo.UpdateWithTransaction(tx, business); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error updating business: %w", err)
	}

	// Update user if needed
	if hasUserUpdates {
		user := &models.User{ID: business.UserID}
		if name, ok := userUpdates["name"].(string); ok {
			user.Name = name
		}
		if email, ok := userUpdates["email"].(string); ok {
			user.Email = email
		}
		if phone, ok := userUpdates["phone"].(string); ok {
			user.Phone = phone
		}
		if password, ok := userUpdates["password"].(string); ok {
			user.Password = password
		}
		if status, ok := userUpdates["status"].(int); ok {
			user.Status = status
		} else if statusF, ok := userUpdates["status"].(float64); ok {
			user.Status = int(statusF)
		}
		if err := s.userRepo.UpdateUser(user); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error updating user: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	businessResponse := s.toBusinessResponse(*business)
	return &businessResponse, nil
}

func (s *businessService) DeleteBusiness(id uint) error {
	if id == 0 {
		return errors.New("invalid business ID")
	}

	business, err := s.businessRepo.GetByID(id)
	if err != nil {
		return errors.New("business not found")
	}

	// Start transaction
	tx := s.businessRepo.BeginTransaction()

	// Delete business first
	if err := s.businessRepo.Delete(id); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting business: %w", err)
	}

	// Delete associated user account
	if err := s.userRepo.Delete(business.UserID); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting user account: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *businessService) GetActiveBusinesses() ([]models.BusinessResponse, error) {
	businesses, err := s.businessRepo.GetActiveBusinesses()
	if err != nil {
		return nil, fmt.Errorf("error fetching active businesses: %w", err)
	}

	var businessResponses []models.BusinessResponse
	for _, business := range businesses {
		businessResponses = append(businessResponses, s.toBusinessResponse(business))
	}

	return businessResponses, nil
}

func (s *businessService) GetInactiveBusinesses() ([]models.BusinessResponse, error) {
	businesses, err := s.businessRepo.GetInactiveBusinesses()
	if err != nil {
		return nil, fmt.Errorf("error fetching inactive businesses: %w", err)
	}

	var businessResponses []models.BusinessResponse
	for _, business := range businesses {
		businessResponses = append(businessResponses, s.toBusinessResponse(business))
	}

	return businessResponses, nil
}

func (s *businessService) ChangeBusinessStatus(businessID uint, status int) error {
	if businessID == 0 {
		return errors.New("invalid business ID")
	}

	if status < 0 || status > 1 {
		return errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
	}

	business, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return errors.New("business not found")
	}

	// Start transaction
	tx := s.businessRepo.BeginTransaction()

	// Update business status
	if err := s.businessRepo.UpdateBusinessStatus(businessID, status); err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating business status: %w", err)
	}

	// Update associated user status
	if err := s.userRepo.UpdateUserStatus(business.UserID, status); err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating user status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *businessService) AssignPackage(businessID, packageID uint) error {
	if businessID == 0 || packageID == 0 {
		return errors.New("invalid business ID or package ID")
	}

	// Check if business exists
	_, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return errors.New("business not found")
	}

	// Check if package exists
	_, err = s.packageRepo.GetByID(packageID)
	if err != nil {
		return errors.New("package not found")
	}

	if err := s.businessRepo.AssignPackage(businessID, packageID); err != nil {
		return fmt.Errorf("error assigning package: %w", err)
	}

	return nil
}

func (s *businessService) RemovePackage(businessID uint) error {
	if businessID == 0 {
		return errors.New("invalid business ID")
	}

	_, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return errors.New("business not found")
	}

	if err := s.businessRepo.RemovePackage(businessID); err != nil {
		return fmt.Errorf("error removing package: %w", err)
	}

	return nil
}

func (s *businessService) GetBusinessesByPackage(packageID uint) ([]models.BusinessResponse, error) {
	if packageID == 0 {
		return nil, errors.New("invalid package ID")
	}

	businesses, err := s.businessRepo.GetBusinessesByPackage(packageID)
	if err != nil {
		return nil, fmt.Errorf("error fetching businesses by package: %w", err)
	}

	var businessResponses []models.BusinessResponse
	for _, business := range businesses {
		businessResponses = append(businessResponses, s.toBusinessResponse(business))
	}

	return businessResponses, nil
}

func (s *businessService) GetBusinessesWithoutPackage() ([]models.BusinessResponse, error) {
	businesses, err := s.businessRepo.GetBusinessesWithoutPackage()
	if err != nil {
		return nil, fmt.Errorf("error fetching businesses without package: %w", err)
	}

	var businessResponses []models.BusinessResponse
	for _, business := range businesses {
		businessResponses = append(businessResponses, s.toBusinessResponse(business))
	}

	return businessResponses, nil
}

func (s *businessService) BusinessEmailExists(email string, excludeBusinessID ...uint) (bool, error) {
	if email == "" {
		return false, errors.New("email cannot be empty")
	}

	return s.businessRepo.BusinessEmailExists(email, excludeBusinessID...)
}

func (s *businessService) BusinessNameExists(name string, excludeBusinessID ...uint) (bool, error) {
	if name == "" {
		return false, errors.New("business name cannot be empty")
	}

	return s.businessRepo.BusinessNameExists(name, excludeBusinessID...)
}

func (s *businessService) GetBusinessStats() (map[string]interface{}, error) {
	stats, err := s.businessRepo.GetBusinessStats()
	if err != nil {
		return nil, fmt.Errorf("error getting business statistics: %w", err)
	}

	return stats, nil
}

func (s *businessService) GetLocationStats() (map[string]int64, error) {
	stats, err := s.businessRepo.GetLocationStats()
	if err != nil {
		return nil, fmt.Errorf("error getting location statistics: %w", err)
	}

	return stats, nil
}

func (s *businessService) GetPackageDistribution() (map[string]int64, error) {
	stats, err := s.businessRepo.GetPackageDistribution()
	if err != nil {
		return nil, fmt.Errorf("error getting package distribution: %w", err)
	}

	return stats, nil
}

func (s *businessService) SearchBusinesses(searchTerm string, limit int) ([]models.BusinessResponse, error) {
	if searchTerm == "" {
		return []models.BusinessResponse{}, nil
	}

	businesses, err := s.businessRepo.SearchBusinesses(searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("error searching businesses: %w", err)
	}

	var businessResponses []models.BusinessResponse
	for _, business := range businesses {
		businessResponses = append(businessResponses, s.toBusinessResponse(business))
	}

	return businessResponses, nil
}

func (s *businessService) BulkUpdateBusinessStatus(businessIDs []uint, status int) error {
	if len(businessIDs) == 0 {
		return errors.New("no business IDs provided")
	}

	if status < 0 || status > 1 {
		return errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
	}

	// Validate that all businesses exist and get their user IDs
	var userIDs []uint
	for _, businessID := range businessIDs {
		business, err := s.businessRepo.GetByID(businessID)
		if err != nil {
			return fmt.Errorf("business with ID %d not found", businessID)
		}
		userIDs = append(userIDs, business.UserID)
	}

	// Start transaction
	tx := s.businessRepo.BeginTransaction()

	// Update business statuses
	if err := s.businessRepo.BulkUpdateStatus(businessIDs, status); err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating business statuses: %w", err)
	}

	// Update associated user statuses
	if err := s.userRepo.BulkUpdateStatus(userIDs, status); err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating user statuses: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *businessService) BulkAssignPackage(businessIDs []uint, packageID uint) error {
	if len(businessIDs) == 0 {
		return errors.New("no business IDs provided")
	}

	if packageID == 0 {
		return errors.New("invalid package ID")
	}

	// Check if package exists
	_, err := s.packageRepo.GetByID(packageID)
	if err != nil {
		return errors.New("package not found")
	}

	// Validate that all businesses exist
	for _, businessID := range businessIDs {
		if _, err := s.businessRepo.GetByID(businessID); err != nil {
			return fmt.Errorf("business with ID %d not found", businessID)
		}
	}

	if err := s.businessRepo.BulkAssignPackage(businessIDs, packageID); err != nil {
		return fmt.Errorf("error assigning package to businesses: %w", err)
	}

	return nil
}

func (s *businessService) GetBusinessesByLocation(location string) ([]models.BusinessResponse, error) {
	if location == "" {
		return nil, errors.New("location cannot be empty")
	}

	businesses, err := s.businessRepo.GetBusinessesByLocation(location)
	if err != nil {
		return nil, fmt.Errorf("error fetching businesses by location: %w", err)
	}

	var businessResponses []models.BusinessResponse
	for _, business := range businesses {
		businessResponses = append(businessResponses, s.toBusinessResponse(business))
	}

	return businessResponses, nil
}

func (s *businessService) GetBusinessLocations() ([]string, error) {
	locations, err := s.businessRepo.GetBusinessLocations()
	if err != nil {
		return nil, fmt.Errorf("error fetching business locations: %w", err)
	}

	return locations, nil
}

// Helper methods

func (s *businessService) toBusinessResponse(business models.Business) models.BusinessResponse {
	return models.BusinessResponse{
		ID:        business.ID,
		Name:      business.Name,
		UserID:    business.UserID,
		OwnerName: business.OwnerName,
		PackageID: business.PackageID,
		Email:     business.Email,
		Phone:     business.Phone,
		Location:  business.Location,
		Status:    business.Status,
		CreatedOn: business.CreatedOn,
	}
}

func (s *businessService) toBusinessResponseWithRelations(business models.Business) models.BusinessResponse {
	response := models.BusinessResponse{
		ID:        business.ID,
		Name:      business.Name,
		UserID:    business.UserID,
		OwnerName: business.OwnerName,
		PackageID: business.PackageID,
		Email:     business.Email,
		Phone:     business.Phone,
		Location:  business.Location,
		Status:    business.Status,
		CreatedOn: business.CreatedOn,
	}

	// Add user relation if loaded
	if business.User.ID != 0 {
		userResponse := models.UserResponse{
			ID:        business.User.ID,
			Name:      business.User.Name,
			Email:     business.User.Email,
			Phone:     business.User.Phone,
			Role:      business.User.Role,
			Status:    business.User.Status,
			CreatedOn: business.User.CreatedOn,
		}
		response.User = &userResponse
	}

	// Add package relation if loaded
	if business.Package != nil && business.Package.ID != 0 {
		packageResponse := models.PackageResponse{
			ID:               business.Package.ID,
			Name:             business.Package.Name,
			Price:            business.Package.Price,
			ValidationPeriod: business.Package.ValidationPeriod,
			Description:      business.Package.Description,
			Status:           business.Package.Status,
			CreatedOn:        business.Package.CreatedOn,
		}
		response.Package = &packageResponse
	}

	return response
}
