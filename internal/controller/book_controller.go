package controller

import (
	"errors"
	"net/http"

	"books-api/internal/dto"
	"books-api/internal/middleware"
	"books-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BookController struct {
	svc     *service.BookService
	authSvc *service.AuthService
}

func NewBookController(svc *service.BookService, authSvc *service.AuthService) *BookController {
	return &BookController{svc: svc, authSvc: authSvc}
}

func (c *BookController) RegisterRoutes(r *gin.Engine) {
	books := r.Group("/books")
	books.Use(middleware.AuthRequired(c.authSvc))
	{
		books.GET("", c.GetAll)
		books.GET("/:id", c.GetByID)
		books.POST("", c.Create)
		books.PUT("/:id", c.Update)
		books.PATCH("/:id", c.Patch)
		books.DELETE("/:id", c.Delete)
	}
}

// GetAll godoc
// @Summary      Список книг
// @Description  Возвращает список книг текущего пользователя с пагинацией. Удалённые книги не включаются.
// @Tags         books
// @Produce      json
// @Param        page   query  int  false  "Номер страницы"        default(1)
// @Param        limit  query  int  false  "Количество на странице" default(10)
// @Success      200    {object}  service.PaginatedBooksResponse
// @Failure      400    {object}  dto.ErrorResponse
// @Failure      401    {object}  dto.ErrorResponse
// @Failure      500    {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /books [get]
func (c *BookController) GetAll(ctx *gin.Context) {
	var p dto.PaginationDTO
	if err := ctx.ShouldBindQuery(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p.SetDefaults()

	result, err := c.svc.GetAll(middleware.GetUserID(ctx), &p)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// GetByID godoc
// @Summary      Получить книгу по ID
// @Description  Возвращает книгу по UUID. Возвращает 404 если книга не найдена, удалена или принадлежит другому пользователю.
// @Tags         books
// @Produce      json
// @Param        id   path  string  true  "UUID книги"
// @Success      200  {object}  models.Book
// @Failure      400  {object}  dto.ErrorResponse  "Неверный формат UUID"
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /books/{id} [get]
func (c *BookController) GetByID(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		return
	}

	book, err := c.svc.GetByID(middleware.GetUserID(ctx), id)
	if err != nil {
		handleBookError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, book)
}

// Create godoc
// @Summary      Создать книгу
// @Description  Создаёт новую книгу и привязывает её к текущему пользователю.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateBookDTO  true  "Данные книги"
// @Success      201   {object}  models.Book
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /books [post]
func (c *BookController) Create(ctx *gin.Context) {
	var body dto.CreateBookDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := c.svc.Create(middleware.GetUserID(ctx), &body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusCreated, book)
}

// Update godoc
// @Summary      Полное обновление книги
// @Description  Полностью заменяет все поля книги. Все поля обязательны.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id    path  string             true  "UUID книги"
// @Param        body  body  dto.UpdateBookDTO  true  "Новые данные книги"
// @Success      200   {object}  models.Book
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /books/{id} [put]
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

	book, err := c.svc.Update(middleware.GetUserID(ctx), id, &body)
	if err != nil {
		handleBookError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, book)
}

// Patch godoc
// @Summary      Частичное обновление книги
// @Description  Обновляет только переданные поля. Непереданные поля остаются без изменений.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id    path  string            true  "UUID книги"
// @Param        body  body  dto.PatchBookDTO  true  "Поля для обновления"
// @Success      200   {object}  models.Book
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /books/{id} [patch]
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

	book, err := c.svc.Patch(middleware.GetUserID(ctx), id, &body)
	if err != nil {
		handleBookError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, book)
}

// Delete godoc
// @Summary      Мягкое удаление книги
// @Description  Помечает книгу как удалённую (Soft Delete). Запись остаётся в БД, но не возвращается в запросах.
// @Tags         books
// @Param        id  path  string  true  "UUID книги"
// @Success      204  "Книга удалена"
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /books/{id} [delete]
func (c *BookController) Delete(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		return
	}

	if err := c.svc.Delete(middleware.GetUserID(ctx), id); err != nil {
		handleBookError(ctx, err)
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

func handleBookError(ctx *gin.Context, err error) {
	if errors.Is(err, service.ErrNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
