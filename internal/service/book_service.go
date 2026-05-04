package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"books-api/internal/cache"
	"books-api/internal/config"
	"books-api/internal/dto"
	"books-api/internal/models"
	"books-api/internal/repository"
)

var (
	ErrNotFound  = errors.New("book not found")
	ErrForbidden = errors.New("forbidden")
)

type BookService struct {
	repo  *repository.BookRepository
	cache *cache.CacheService
	cfg   *config.Config
}

func NewBookService(repo *repository.BookRepository, cache *cache.CacheService, cfg *config.Config) *BookService {
	return &BookService{repo: repo, cache: cache, cfg: cfg}
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

func (s *BookService) GetAll(userID uuid.UUID, p *dto.PaginationDTO) (*PaginatedBooksResponse, error) {
	ctx := context.Background()
	key := bookListKey(userID, p.Page, p.Limit)
	ttl := time.Duration(s.cfg.CacheTTL) * time.Second

	if val, err := s.cache.Get(ctx, key); err == nil {
		var resp PaginatedBooksResponse
		if json.Unmarshal([]byte(val), &resp) == nil {
			return &resp, nil
		}
	}

	result, err := s.repo.FindAll(userID, p)
	if err != nil {
		return nil, err
	}

	totalPages := int(result.Total) / p.Limit
	if int(result.Total)%p.Limit != 0 {
		totalPages++
	}

	resp := &PaginatedBooksResponse{
		Data: result.Books,
		Meta: PaginationMeta{
			Total:      result.Total,
			Page:       p.Page,
			Limit:      p.Limit,
			TotalPages: totalPages,
		},
	}

	if data, err := json.Marshal(resp); err == nil {
		_ = s.cache.Set(ctx, key, string(data), ttl)
	}
	return resp, nil
}

func (s *BookService) GetByID(userID, bookID uuid.UUID) (*models.Book, error) {
	ctx := context.Background()
	key := bookItemKey(bookID)
	ttl := time.Duration(s.cfg.CacheTTL) * time.Second

	if val, err := s.cache.Get(ctx, key); err == nil {
		var book models.Book
		if json.Unmarshal([]byte(val), &book) == nil {
			if book.UserID != userID {
				return nil, ErrNotFound
			}
			return &book, nil
		}
	}

	book, err := s.repo.FindByID(bookID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if book.UserID != userID {
		return nil, ErrNotFound
	}

	if data, err := json.Marshal(book); err == nil {
		_ = s.cache.Set(ctx, key, string(data), ttl)
	}
	return book, nil
}

func (s *BookService) Create(userID uuid.UUID, d *dto.CreateBookDTO) (*models.Book, error) {
	book := &models.Book{
		UserID:      userID,
		Title:       d.Title,
		Author:      d.Author,
		Description: d.Description,
		Year:        d.Year,
	}
	if err := s.repo.Create(book); err != nil {
		return nil, err
	}
	_ = s.cache.DelByPattern(context.Background(), bookListPattern(userID))
	return book, nil
}

func (s *BookService) Update(userID, bookID uuid.UUID, d *dto.UpdateBookDTO) (*models.Book, error) {
	book, err := s.GetByID(userID, bookID)
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
	s.invalidateBook(context.Background(), userID, bookID)
	return book, nil
}

func (s *BookService) Patch(userID, bookID uuid.UUID, d *dto.PatchBookDTO) (*models.Book, error) {
	book, err := s.GetByID(userID, bookID)
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
	s.invalidateBook(context.Background(), userID, bookID)
	return book, nil
}

func (s *BookService) Delete(userID, bookID uuid.UUID) error {
	if _, err := s.GetByID(userID, bookID); err != nil {
		return err
	}
	if err := s.repo.Delete(bookID); err != nil {
		return err
	}
	s.invalidateBook(context.Background(), userID, bookID)
	return nil
}

func (s *BookService) invalidateBook(ctx context.Context, userID, bookID uuid.UUID) {
	_ = s.cache.Del(ctx, bookItemKey(bookID))
	_ = s.cache.DelByPattern(ctx, bookListPattern(userID))
}

// --- Cache key helpers ---

func bookListKey(userID uuid.UUID, page, limit int) string {
	return fmt.Sprintf("wp:books:list:user:%s:page:%d:limit:%d", userID, page, limit)
}

func bookListPattern(userID uuid.UUID) string {
	return fmt.Sprintf("wp:books:list:user:%s:*", userID)
}

func bookItemKey(bookID uuid.UUID) string {
	return fmt.Sprintf("wp:books:item:%s", bookID)
}
