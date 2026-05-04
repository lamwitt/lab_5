package service

import (
	"errors"

	"books-api/internal/dto"
	"books-api/internal/models"
	"books-api/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("book not found")

type BookService struct {
	repo *repository.BookRepository
}

func NewBookService(repo *repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedBooksResponse struct {
	Data []models.Book  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func (s *BookService) GetAll(p *dto.PaginationDTO) (*PaginatedBooksResponse, error) {
	result, err := s.repo.FindAll(p)
	if err != nil {
		return nil, err
	}

	totalPages := int(result.Total) / p.Limit
	if int(result.Total)%p.Limit != 0 {
		totalPages++
	}

	return &PaginatedBooksResponse{
		Data: result.Books,
		Meta: PaginationMeta{
			Total:      result.Total,
			Page:       p.Page,
			Limit:      p.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *BookService) GetByID(id uuid.UUID) (*models.Book, error) {
	book, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return book, nil
}

func (s *BookService) Create(d *dto.CreateBookDTO) (*models.Book, error) {
	book := &models.Book{
		Title:       d.Title,
		Author:      d.Author,
		Description: d.Description,
		Year:        d.Year,
	}
	if err := s.repo.Create(book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *BookService) Update(id uuid.UUID, d *dto.UpdateBookDTO) (*models.Book, error) {
	book, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	book.Title = d.Title
	book.Author = d.Author
	book.Description = d.Description
	book.Year = d.Year

	if err := s.repo.Save(book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *BookService) Patch(id uuid.UUID, d *dto.PatchBookDTO) (*models.Book, error) {
	book, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if d.Title != nil {
		book.Title = *d.Title
	}
	if d.Author != nil {
		book.Author = *d.Author
	}
	if d.Description != nil {
		book.Description = *d.Description
	}
	if d.Year != nil {
		book.Year = *d.Year
	}

	if err := s.repo.Save(book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *BookService) Delete(id uuid.UUID) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
