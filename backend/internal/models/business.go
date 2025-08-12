package models

import (
	"time"
)

type Business struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;not null"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	OwnerName string    `json:"owner_name" gorm:"not null"`
	PackageID *uint     `json:"package_id" gorm:"default:null"` // nullable
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Phone     string    `json:"phone"`
	Location  string    `json:"location"`
	Password  string    `json:"-" gorm:"not null"`
	Status    int       `json:"status" gorm:"not null;default:1"` // 1=active, 0=inactive
	CreatedOn time.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn time.Time `json:"updated_on" gorm:"column:updated_on;autoUpdateTime"`

	// Relationships
	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Package *Package `json:"package,omitempty" gorm:"foreignKey:PackageID"`
}

// TableName overrides the table name
func (Business) TableName() string {
	return "business"
}

type BusinessResponse struct {
	ID        uint             `json:"id"`
	Name      string           `json:"name"`
	Slug      string           `json:"slug"`
	UserID    uint             `json:"user_id"`
	OwnerName string           `json:"owner_name"`
	PackageID *uint            `json:"package_id"`
	Email     string           `json:"email"`
	Phone     string           `json:"phone"`
	Location  string           `json:"location"`
	Status    int              `json:"status"`
	CreatedOn time.Time        `json:"created_on"`
	User      *UserResponse    `json:"user,omitempty"`
	Package   *PackageResponse `json:"package,omitempty"`
}

type CreateBusinessRequest struct {
	Name      string `json:"name" binding:"required"`
	Slug      string `json:"slug" binding:"required"`
	OwnerName string `json:"owner_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
	Password  string `json:"password" binding:"required,min=6"`
	PackageID *uint  `json:"package_id"` // optional
}

type UpdateBusinessRequest struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	OwnerName string `json:"owner_name"`
	Email     string `json:"email" binding:"omitempty,email"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
	Password  string `json:"password" binding:"omitempty,min=6"`
	PackageID *uint  `json:"package_id"`
	Status    *int   `json:"status"` // pointer to allow null/zero values
}

type AssignPackageRequest struct {
	PackageID uint `json:"package_id" binding:"required"`
}
