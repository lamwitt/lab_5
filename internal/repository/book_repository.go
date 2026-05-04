package repository

import (
	"books-api/internal/dto"
	"books-api/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{db: db}
}

type PaginatedResult struct {
	Books []models.Book
	Total int64
}

func (r *BookRepository) FindAll(userID uuid.UUID, p *dto.PaginationDTO) (*PaginatedResult, error) {
	var books []models.Book
	var total int64

	offset := (p.Page - 1) * p.Limit
	q := r.db.Model(&models.Book{}).Where("user_id = ?", userID)

	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := q.Offset(offset).Limit(p.Limit).Find(&books).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult{Books: books, Total: total}, nil
}

func (r *BookRepository) FindByID(id uuid.UUID) (*models.Book, error) {
	var book models.Book
	if err := r.db.First(&book, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *BookRepository) Create(book *models.Book) error {
	return r.db.Create(book).Error
}

func (r *BookRepository) Save(book *models.Book) error {
	return r.db.Save(book).Error
}

func (r *BookRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Book{}, "id = ?", id).Error
}
