package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// UserRole represents the different user roles in the system
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleBusiness UserRole = "business"
	RoleTeacher  UserRole = "teacher"
	RoleStudent  UserRole = "student"
)

// IsValid checks if the user role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleBusiness, RoleTeacher, RoleStudent:
		return true
	default:
		return false
	}
}

// String returns the string representation of UserRole
func (r UserRole) String() string {
	return string(r)
}

// Scan implements the sql.Scanner interface - FIXED VERSION
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = RoleStudent
		return nil
	}

	switch s := value.(type) {
	case string:
		*r = UserRole(s)
		return nil
	case []byte:
		*r = UserRole(string(s))
		return nil
	default:
		// Return error instead of trying to handle unknown types
		return fmt.Errorf("cannot scan %T into UserRole", value)
	}
}

// Value implements the driver.Valuer interface - FIXED VERSION
func (r UserRole) Value() (driver.Value, error) {
	if r == "" {
		return string(RoleStudent), nil
	}
	return string(r), nil
}

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"` // Changed from unique to uniqueIndex
	Phone     string    `json:"phone"`
	Password  string    `json:"-" gorm:"not null"`
	Role      UserRole  `json:"role" gorm:"type:varchar(20);not null;default:'student'"` // Added not null
	Status    int       `json:"status" gorm:"not null;default:1"`                        // Added not null
	CreatedOn time.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn time.Time `json:"updated_on" gorm:"column:updated_on;autoUpdateTime"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      UserRole  `json:"role"`
	Status    int       `json:"status"`
	CreatedOn time.Time `json:"created_on"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateUserRequest struct {
	Name     string   `json:"name" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Phone    string   `json:"phone"`
	Password string   `json:"password" binding:"required,min=6"`
	Role     UserRole `json:"role" binding:"omitempty,oneof=admin business teacher student"`
}
