package controller

import (
	"errors"
	"net/http"

	"books-api/internal/dto"
	"books-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BookController struct {
	svc *service.BookService
}

func NewBookController(svc *service.BookService) *BookController {
	return &BookController{svc: svc}
}

func (c *BookController) RegisterRoutes(r *gin.Engine) {
	books := r.Group("/books")
	{
		books.GET("", c.GetAll)
		books.GET("/:id", c.GetByID)
		books.POST("", c.Create)
		books.PUT("/:id", c.Update)
		books.PATCH("/:id", c.Patch)
		books.DELETE("/:id", c.Delete)
	}
}

func (c *BookController) GetAll(ctx *gin.Context) {
	var p dto.PaginationDTO
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p.SetDefaults()

	result, err := c.svc.GetAll(&p)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *BookController) GetByID(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		return
	}

	book, err := c.svc.GetByID(id)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, book)
}

func (c *BookController) Create(ctx *gin.Context) {
	var body dto.CreateBookDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := c.svc.Create(&body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusCreated, book)
}

func (c *BookController) Update(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		return
	}

	var body dto.UpdateBookDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := c.svc.Update(id, &body)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, book)
}

func (c *BookController) Patch(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		return
	}

	var body dto.PatchBookDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := c.svc.Patch(id, &body)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, book)
}

func (c *BookController) Delete(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		return
	}

	if err := c.svc.Delete(id); err != nil {
		handleServiceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func parseID(ctx *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return uuid.Nil, err
	}
	return id, nil
}

func handleServiceError(ctx *gin.Context, err error) {
	if errors.Is(err, service.ErrNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
