package services

import (
	"backend/internal/models"
	"backend/internal/repository"
	"errors"
	"fmt"
)

type PackageService interface {
	CreatePackage(req models.CreatePackageRequest) (*models.PackageResponse, error)
	GetPackages(filters repository.PackageFilters) ([]models.PackageResponse, int64, error)
	GetPackageByID(id uint) (*models.PackageResponse, error)
	UpdatePackage(id uint, updates map[string]interface{}) (*models.PackageResponse, error)
	DeletePackage(id uint) error
	GetActivePackages() ([]models.PackageResponse, error)
	GetInactivePackages() ([]models.PackageResponse, error)
	ChangePackageStatus(packageID uint, status int) error
	PackageNameExists(name string, excludePackageID ...uint) (bool, error)
	GetPackageStats() (map[string]interface{}, error)
	GetPriceStatistics() (map[string]float64, error)
	GetPackagesByPriceRange(minPrice, maxPrice float64) ([]models.PackageResponse, error)
	BulkUpdatePackageStatus(packageIDs []uint, status int) error
	SearchPackages(searchTerm string, limit int) ([]models.PackageResponse, error)
}

type packageService struct {
	repo repository.PackageRepository
}

func NewPackageService(repo repository.PackageRepository) PackageService {
	return &packageService{
		repo: repo,
	}
}

func (s *packageService) CreatePackage(req models.CreatePackageRequest) (*models.PackageResponse, error) {
	// Check if package name already exists
	exists, err := s.repo.PackageNameExists(req.Name)
	if err != nil {
		return nil, fmt.Errorf("error checking package name existence: %w", err)
	}
	if exists {
		return nil, errors.New("package name already exists")
	}

	// Validate price
	if req.Price < 0 {
		return nil, errors.New("package price cannot be negative")
	}

	// Validate validation period
	if req.ValidationPeriod <= 0 {
		return nil, errors.New("validation period must be greater than 0 days")
	}

	pkg := &models.Package{
		Name:             req.Name,
		Price:            req.Price,
		ValidationPeriod: req.ValidationPeriod,
		Description:      req.Description,
		Status:           1, // Active by default
	}

	if err := s.repo.Create(pkg); err != nil {
		return nil, fmt.Errorf("error creating package: %w", err)
	}

	packageResponse := s.toPackageResponse(*pkg)
	return &packageResponse, nil
}

func (s *packageService) GetPackages(filters repository.PackageFilters) ([]models.PackageResponse, int64, error) {
	// Set default pagination if not provided
	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	packages, total, err := s.repo.GetAll(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching packages: %w", err)
	}

	var packageResponses []models.PackageResponse
	for _, pkg := range packages {
		packageResponses = append(packageResponses, s.toPackageResponse(pkg))
	}

	return packageResponses, total, nil
}

func (s *packageService) GetPackageByID(id uint) (*models.PackageResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid package ID")
	}

	pkg, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("package not found")
	}

	packageResponse := s.toPackageResponse(*pkg)
	return &packageResponse, nil
}

func (s *packageService) UpdatePackage(id uint, updates map[string]interface{}) (*models.PackageResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid package ID")
	}

	pkg, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("package not found")
	}

	// Track if any updates were made
	hasUpdates := false

	// Apply updates with validation
	if name, ok := updates["name"].(string); ok && name != "" {
		// Check if name already exists for another package
		exists, err := s.repo.PackageNameExists(name, pkg.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking package name existence: %w", err)
		}
		if exists {
			return nil, errors.New("package name already exists")
		}
		pkg.Name = name
		hasUpdates = true
	}

	if price, ok := updates["price"].(float64); ok {
		if price < 0 {
			return nil, errors.New("package price cannot be negative")
		}
		pkg.Price = price
		hasUpdates = true
	}

	if validationPeriod, ok := updates["validation_period"].(float64); ok {
		periodInt := int(validationPeriod)
		if periodInt <= 0 {
			return nil, errors.New("validation period must be greater than 0 days")
		}
		pkg.ValidationPeriod = periodInt
		hasUpdates = true
	}

	if validationPeriod, ok := updates["validation_period"].(int); ok {
		if validationPeriod <= 0 {
			return nil, errors.New("validation period must be greater than 0 days")
		}
		pkg.ValidationPeriod = validationPeriod
		hasUpdates = true
	}

	if description, ok := updates["description"].(string); ok {
		pkg.Description = description
		hasUpdates = true
	}

	if status, ok := updates["status"].(int); ok {
		if status < 0 || status > 1 {
			return nil, errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
		}
		pkg.Status = status
		hasUpdates = true
	}

	// Handle float64 status (from JSON)
	if status, ok := updates["status"].(float64); ok {
		statusInt := int(status)
		if statusInt < 0 || statusInt > 1 {
			return nil, errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
		}
		pkg.Status = statusInt
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, errors.New("no valid updates provided")
	}

	if err := s.repo.Update(pkg); err != nil {
		return nil, fmt.Errorf("error updating package: %w", err)
	}

	packageResponse := s.toPackageResponse(*pkg)
	return &packageResponse, nil
}

func (s *packageService) DeletePackage(id uint) error {
	if id == 0 {
		return errors.New("invalid package ID")
	}

	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("package not found")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("error deleting package: %w", err)
	}

	return nil
}

func (s *packageService) GetActivePackages() ([]models.PackageResponse, error) {
	packages, err := s.repo.GetActivePackages()
	if err != nil {
		return nil, fmt.Errorf("error fetching active packages: %w", err)
	}

	var packageResponses []models.PackageResponse
	for _, pkg := range packages {
		packageResponses = append(packageResponses, s.toPackageResponse(pkg))
	}

	return packageResponses, nil
}

func (s *packageService) GetInactivePackages() ([]models.PackageResponse, error) {
	packages, err := s.repo.GetInactivePackages()
	if err != nil {
		return nil, fmt.Errorf("error fetching inactive packages: %w", err)
	}

	var packageResponses []models.PackageResponse
	for _, pkg := range packages {
		packageResponses = append(packageResponses, s.toPackageResponse(pkg))
	}

	return packageResponses, nil
}

func (s *packageService) ChangePackageStatus(packageID uint, status int) error {
	if packageID == 0 {
		return errors.New("invalid package ID")
	}

	if status < 0 || status > 1 {
		return errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
	}

	_, err := s.repo.GetByID(packageID)
	if err != nil {
		return errors.New("package not found")
	}

	if err := s.repo.UpdatePackageStatus(packageID, status); err != nil {
		return fmt.Errorf("error updating package status: %w", err)
	}

	return nil
}

func (s *packageService) PackageNameExists(name string, excludePackageID ...uint) (bool, error) {
	if name == "" {
		return false, errors.New("package name cannot be empty")
	}

	return s.repo.PackageNameExists(name, excludePackageID...)
}

func (s *packageService) GetPackageStats() (map[string]interface{}, error) {
	stats, err := s.repo.GetPackageStats()
	if err != nil {
		return nil, fmt.Errorf("error getting package statistics: %w", err)
	}

	return stats, nil
}

func (s *packageService) GetPriceStatistics() (map[string]float64, error) {
	stats, err := s.repo.GetPriceStatistics()
	if err != nil {
		return nil, fmt.Errorf("error getting price statistics: %w", err)
	}

	return stats, nil
}

func (s *packageService) GetPackagesByPriceRange(minPrice, maxPrice float64) ([]models.PackageResponse, error) {
	if minPrice < 0 || maxPrice < 0 {
		return nil, errors.New("price values cannot be negative")
	}

	if minPrice > maxPrice && maxPrice > 0 {
		return nil, errors.New("minimum price cannot be greater than maximum price")
	}

	packages, err := s.repo.GetPackagesByPriceRange(minPrice, maxPrice)
	if err != nil {
		return nil, fmt.Errorf("error fetching packages by price range: %w", err)
	}

	var packageResponses []models.PackageResponse
	for _, pkg := range packages {
		packageResponses = append(packageResponses, s.toPackageResponse(pkg))
	}

	return packageResponses, nil
}

func (s *packageService) BulkUpdatePackageStatus(packageIDs []uint, status int) error {
	if len(packageIDs) == 0 {
		return errors.New("no package IDs provided")
	}

	if status < 0 || status > 1 {
		return errors.New("invalid status value. Must be 0 (inactive) or 1 (active)")
	}

	// Validate that all packages exist
	for _, packageID := range packageIDs {
		if _, err := s.repo.GetByID(packageID); err != nil {
			return fmt.Errorf("package with ID %d not found", packageID)
		}
	}

	if err := s.repo.BulkUpdateStatus(packageIDs, status); err != nil {
		return fmt.Errorf("error updating package statuses: %w", err)
	}

	return nil
}

// Helper methods

func (s *packageService) GetPackagesByValidationPeriod(minDays, maxDays int) ([]models.PackageResponse, error) {
	if minDays < 0 || maxDays < 0 {
		return nil, errors.New("validation period values cannot be negative")
	}

	if minDays > maxDays && maxDays > 0 {
		return nil, errors.New("minimum validation period cannot be greater than maximum validation period")
	}

	packages, err := s.repo.GetPackagesByValidationPeriod(minDays, maxDays)
	if err != nil {
		return nil, fmt.Errorf("error fetching packages by validation period: %w", err)
	}

	var packageResponses []models.PackageResponse
	for _, pkg := range packages {
		packageResponses = append(packageResponses, s.toPackageResponse(pkg))
	}

	return packageResponses, nil
}

func (s *packageService) SearchPackages(searchTerm string, limit int) ([]models.PackageResponse, error) {
	if searchTerm == "" {
		return []models.PackageResponse{}, nil
	}

	packages, err := s.repo.SearchPackages(searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("error searching packages: %w", err)
	}

	var packageResponses []models.PackageResponse
	for _, pkg := range packages {
		packageResponses = append(packageResponses, s.toPackageResponse(pkg))
	}

	return packageResponses, nil
}

func (s *packageService) GetRecentPackages(limit int) ([]models.PackageResponse, error) {
	packages, err := s.repo.GetRecentPackages(limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching recent packages: %w", err)
	}

	var packageResponses []models.PackageResponse
	for _, pkg := range packages {
		packageResponses = append(packageResponses, s.toPackageResponse(pkg))
	}

	return packageResponses, nil
}

func (s *packageService) ValidatePackageData(req models.CreatePackageRequest) error {
	if req.Name == "" {
		return errors.New("package name is required")
	}

	if req.Price < 0 {
		return errors.New("package price cannot be negative")
	}

	if req.ValidationPeriod <= 0 {
		return errors.New("validation period must be greater than 0 days")
	}

	// Additional business logic validations can be added here
	if req.ValidationPeriod > 365 {
		return errors.New("validation period cannot exceed 365 days")
	}

	return nil
}

func (s *packageService) toPackageResponse(pkg models.Package) models.PackageResponse {
	return models.PackageResponse{
		ID:               pkg.ID,
		Name:             pkg.Name,
		Price:            pkg.Price,
		ValidationPeriod: pkg.ValidationPeriod,
		Description:      pkg.Description,
		Status:           pkg.Status,
		CreatedOn:        pkg.CreatedOn,
	}
}
