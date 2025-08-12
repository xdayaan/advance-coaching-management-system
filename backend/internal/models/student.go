package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONB type for PostgreSQL
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

type Student struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name" gorm:"not null"`
	UserID         uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	BusinessID     uint      `json:"business_id" gorm:"not null"`
	GuardianName   string    `json:"guardian_name"`
	GuardianNumber string    `json:"guardian_number"`
	GuardianEmail  string    `json:"guardian_email"`
	Information    JSONB     `json:"information" gorm:"type:jsonb"`
	Status         int       `json:"status" gorm:"not null;default:1"` // 1=active, 0=inactive
	CreatedOn      time.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn      time.Time `json:"updated_on" gorm:"column:updated_on;autoUpdateTime"`

	// Relationships
	User     User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Business Business `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
}

// TableName overrides the table name
func (Student) TableName() string {
	return "student"
}

type StudentResponse struct {
	ID             uint              `json:"id"`
	Name           string            `json:"name"`
	UserID         uint              `json:"user_id"`
	BusinessID     uint              `json:"business_id"`
	GuardianName   string            `json:"guardian_name"`
	GuardianNumber string            `json:"guardian_number"`
	GuardianEmail  string            `json:"guardian_email"`
	Information    JSONB             `json:"information"`
	Status         int               `json:"status"`
	CreatedOn      time.Time         `json:"created_on"`
	UpdatedOn      time.Time         `json:"updated_on"`
	User           *UserResponse     `json:"user,omitempty"`
	Business       *BusinessResponse `json:"business,omitempty"`
}

type CreateStudentRequest struct {
	Name           string `json:"name" binding:"required"`
	UserID         uint   `json:"user_id" binding:"required"`
	BusinessID     uint   `json:"business_id" binding:"required"`
	GuardianName   string `json:"guardian_name"`
	GuardianNumber string `json:"guardian_number"`
	GuardianEmail  string `json:"guardian_email" binding:"omitempty,email"`
	Information    JSONB  `json:"information"`
}

type UpdateStudentRequest struct {
	Name           string `json:"name"`
	GuardianName   string `json:"guardian_name"`
	GuardianNumber string `json:"guardian_number"`
	GuardianEmail  string `json:"guardian_email" binding:"omitempty,email"`
	Information    JSONB  `json:"information"`
	Status         *int   `json:"status"`
}

type StudentStatsResponse struct {
	TotalStudents    int64 `json:"total_students"`
	ActiveStudents   int64 `json:"active_students"`
	InactiveStudents int64 `json:"inactive_students"`
}
