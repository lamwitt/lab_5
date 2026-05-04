package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255"                  json:"email"`
	PasswordHash string         `gorm:"size:255"                                       json:"-"`
	Salt         string         `gorm:"size:255"                                       json:"-"`
	YandexID     *string        `gorm:"size:255;uniqueIndex"                           json:"-"`
	CreatedAt    time.Time      `                                                       json:"createdAt"`
	UpdatedAt    time.Time      `                                                       json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                                          json:"-"`
}
