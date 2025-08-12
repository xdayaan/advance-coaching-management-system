package models

import (
	"time"
)

type Package struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"not null;uniqueIndex"`
	Price            float64   `json:"price" gorm:"not null"`
	ValidationPeriod int       `json:"validation_period" gorm:"not null"`
	Description      string    `json:"description"`
	Status           int       `json:"Status" validate:"required"`
	CreatedOn        time.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn        time.Time `json:"updated_on" gorm:"column:updated_on;autoUpdateTime"`
}

// TableName overrides the table name
func (Package) TableName() string {
	return "packages"
}

type PackageResponse struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	Price            float64   `json:"price"`
	ValidationPeriod int       `json:"validation_period"`
	Description      string    `json:"description"`
	Status           int       `json:"status"`
	CreatedOn        time.Time `json:"created_on"`
}

type CreatePackageRequest struct {
	Name             string  `json:"name" binding:"required"`
	Price            float64 `json:"price" binding:"required,min=0"`
	ValidationPeriod int     `json:"validation_period" binding:"required,min=1"`
	Description      string  `json:"description"`
}

type UpdatePackageRequest struct {
	Name             string  `json:"name"`
	Price            float64 `json:"price" binding:"min=0"`
	ValidationPeriod int     `json:"validation_period" binding:"min=1"`
	Description      string  `json:"description"`
	Status           *int    `json:"status"` // pointer to allow null/zero values
}
