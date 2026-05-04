package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Book struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string         `gorm:"not null;size:255"                              json:"title"`
	Author      string         `gorm:"not null;size:255"                              json:"author"`
	Description string         `gorm:"type:text"                                      json:"description"`
	Year        int            `                                                      json:"year"`
	CreatedAt   time.Time      `                                                      json:"createdAt"`
	UpdatedAt   time.Time      `                                                      json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                          json:"-"`
}
