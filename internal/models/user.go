package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered user
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"` // Hashed password
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	APIKeys []APIKey `gorm:"foreignKey:UserID" json:"api_keys,omitempty"`
}

// APIKey represents a user's API key for accessing the gateway
type APIKey struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Key       string         `gorm:"uniqueIndex;size:64;not null" json:"key"`
	Name      string         `gorm:"size:128" json:"name"`
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Usage stats
	TotalRequests int64 `gorm:"default:0" json:"total_requests"`
	TotalTokens   int64 `gorm:"default:0" json:"total_tokens"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// RequestLog records API requests for auditing
type RequestLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	APIKeyID  uint      `gorm:"index" json:"api_key_id"`
	Method    string    `gorm:"size:10" json:"method"`
	Path      string    `gorm:"size:255" json:"path"`
	Model     string    `gorm:"size:128" json:"model"`
	Tokens    int       `json:"tokens"`
	Duration  int64     `json:"duration_ms"` // Request duration in milliseconds
	Success   bool      `json:"success"`
	Error     string    `gorm:"size:512" json:"error,omitempty"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
