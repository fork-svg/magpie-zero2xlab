package db

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fork-svg/magpie-zero2xlab/config"
	"github.com/fork-svg/magpie-zero2xlab/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init initializes the database connection and creates tables
func Init(cfg *config.Config) error {
	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		return err
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	// Auto migrate tables
	if err := DB.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&models.RequestLog{},
	); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// CreateUser creates a new user
func CreateUser(user *models.User) error {
	return DB.Create(user).Error
}

// GetAPIKey retrieves an API key with user info
func GetAPIKey(key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	result := DB.Preload("User").Where("key = ? AND active = true", key).First(&apiKey)
	if result.Error != nil {
		return nil, result.Error
	}
	return &apiKey, nil
}

// GetUserAPIKeys retrieves all API keys for a user
func GetUserAPIKeys(userID uint) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	result := DB.Where("user_id = ? AND active = true", userID).Find(&apiKeys)
	if result.Error != nil {
		return nil, result.Error
	}
	return apiKeys, nil
}

// CreateAPIKey creates a new API key
func CreateAPIKey(apiKey *models.APIKey) error {
	return DB.Create(apiKey).Error
}

// DeleteAPIKey soft-deletes an API key
func DeleteAPIKey(key string, userID uint) error {
	return DB.Where("key = ? AND user_id = ?", key, userID).Delete(&models.APIKey{}).Error
}

// UpdateAPIKeyUsage updates the usage stats for an API key
func UpdateAPIKeyUsage(apiKeyID uint, tokens int) error {
	return DB.Model(&models.APIKey{}).Where("id = ?", apiKeyID).
		Updates(map[string]interface{}{
			"total_requests": gorm.Expr("total_requests + 1"),
			"total_tokens":   gorm.Expr("total_tokens + ?", tokens),
		}).Error
}

// CreateRequestLog creates a new request log entry
func CreateRequestLog(log *models.RequestLog) error {
	return DB.Create(log).Error
}

// GetUserRequestLogs retrieves request logs for a user's API keys
func GetUserRequestLogs(userID uint, limit int) ([]models.RequestLog, error) {
	var logs []models.RequestLog
	var apiKeyIDs []uint

	// Get user's API key IDs
	if err := DB.Model(&models.APIKey{}).Where("user_id = ?", userID).
		Pluck("id", &apiKeyIDs).Error; err != nil {
		return nil, err
	}

	result := DB.Where("api_key_id IN ?", apiKeyIDs).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs)
	if result.Error != nil {
		return nil, result.Error
	}
	return logs, nil
}
