package models

import (
	"time"
)

type Teacher struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"`
	UserID        uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	BusinessID    uint      `json:"business_id" gorm:"not null"`
	Salary        float64   `json:"salary" gorm:"type:decimal(10,2)"`
	Qualification string    `json:"qualification"`
	Experience    string    `json:"experience"`
	Description   string    `json:"description"`
	Status        int       `json:"status" gorm:"not null;default:1"` // 1=active, 0=inactive
	CreatedOn     time.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn     time.Time `json:"updated_on" gorm:"column:updated_on;autoUpdateTime"`

	// Relationships
	User     User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Business Business `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
}

// TableName overrides the table name
func (Teacher) TableName() string {
	return "teacher"
}

type TeacherResponse struct {
	ID            uint              `json:"id"`
	Name          string            `json:"name"`
	UserID        uint              `json:"user_id"`
	BusinessID    uint              `json:"business_id"`
	Salary        float64           `json:"salary"`
	Qualification string            `json:"qualification"`
	Experience    string            `json:"experience"`
	Description   string            `json:"description"`
	Status        int               `json:"status"`
	CreatedOn     time.Time         `json:"created_on"`
	UpdatedOn     time.Time         `json:"updated_on"`
	User          *UserResponse     `json:"user,omitempty"`
	Business      *BusinessResponse `json:"business,omitempty"`
}

type CreateTeacherRequest struct {
	Name          string  `json:"name" binding:"required"`
	UserID        uint    `json:"user_id" binding:"required"`
	BusinessID    uint    `json:"business_id" binding:"required"`
	Salary        float64 `json:"salary"`
	Qualification string  `json:"qualification"`
	Experience    string  `json:"experience"`
	Description   string  `json:"description"`
}

type UpdateTeacherRequest struct {
	Name          string   `json:"name"`
	Salary        *float64 `json:"salary"`
	Qualification string   `json:"qualification"`
	Experience    string   `json:"experience"`
	Description   string   `json:"description"`
	Status        *int     `json:"status"`
}

type TeacherStatsResponse struct {
	TotalTeachers    int64   `json:"total_teachers"`
	ActiveTeachers   int64   `json:"active_teachers"`
	InactiveTeachers int64   `json:"inactive_teachers"`
	AverageSalary    float64 `json:"average_salary"`
}
