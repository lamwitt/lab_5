package controller

import (
	"net/http"

	"books-api/internal/config"
	"books-api/internal/dto"
	"books-api/internal/middleware"
	"books-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authSvc *service.AuthService
	cfg     *config.Config
}

func NewAuthController(authSvc *service.AuthService, cfg *config.Config) *AuthController {
	return &AuthController{authSvc: authSvc, cfg: cfg}
}

func (c *AuthController) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", c.Register)
		auth.POST("/login", c.Login)
		auth.POST("/refresh", c.Refresh)
		auth.POST("/forgot-password", c.ForgotPassword)
		auth.POST("/reset-password", c.ResetPassword)
		auth.GET("/oauth/:provider", c.OAuthInitiate)
		auth.GET("/oauth/:provider/callback", c.OAuthCallback)

		private := auth.Group("")
		private.Use(middleware.AuthRequired(c.authSvc))
		{
			private.GET("/whoami", c.Whoami)
			private.POST("/logout", c.Logout)
			private.POST("/logout-all", c.LogoutAll)
		}
	}
}

// Register godoc
// @Summary      Регистрация
// @Description  Создаёт новый аккаунт. Пароль хранится в виде bcrypt-хеша с солью.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterDTO     true  "Email и пароль"
// @Success      201   {object}  dto.UserResponseDTO
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse  "Пользователь уже существует"
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var body dto.RegisterDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.authSvc.Register(&body)
	if err != nil {
		if err == service.ErrUserExists {
			ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusCreated, dto.UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// Login godoc
// @Summary      Вход
// @Description  Проверяет учётные данные и устанавливает HttpOnly cookies: access_token (15 мин) и refresh_token (7 дней).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginDTO        true  "Email и пароль"
// @Success      200   {object}  dto.MessageResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse   "Неверные учётные данные"
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var body dto.LoginDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := c.authSvc.Login(&body)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.setTokenCookies(ctx, accessToken, refreshToken)
	ctx.JSON(http.StatusOK, gin.H{"message": "logged in"})
}

// Refresh godoc
// @Summary      Обновление токенов
// @Description  Выдаёт новую пару access/refresh токенов по действующему refresh_token cookie. Старый refresh токен инвалидируется.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dto.MessageResponse
// @Failure      401  {object}  dto.ErrorResponse  "Refresh токен отсутствует или недействителен"
// @Router       /auth/refresh [post]
func (c *AuthController) Refresh(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	accessToken, newRefreshToken, err := c.authSvc.Refresh(refreshToken)
	if err != nil {
		c.clearTokenCookies(ctx)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	c.setTokenCookies(ctx, accessToken, newRefreshToken)
	ctx.JSON(http.StatusOK, gin.H{"message": "tokens refreshed"})
}

// Whoami godoc
// @Summary      Профиль текущего пользователя
// @Description  Возвращает данные авторизованного пользователя. Используется фронтендом для проверки сессии, так как HttpOnly cookies недоступны из JavaScript.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dto.UserResponseDTO
// @Failure      401  {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /auth/whoami [get]
func (c *AuthController) Whoami(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	user, err := c.authSvc.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx.JSON(http.StatusOK, dto.UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// Logout godoc
// @Summary      Выход из текущей сессии
// @Description  Инвалидирует текущую пару токенов в базе данных и очищает cookies.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dto.MessageResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	accessToken, _ := ctx.Cookie("access_token")
	refreshToken, _ := ctx.Cookie("refresh_token")
	c.authSvc.Logout(accessToken, refreshToken)
	c.clearTokenCookies(ctx)
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// LogoutAll godoc
// @Summary      Выход со всех устройств
// @Description  Инвалидирует все токены пользователя в базе данных. Завершает все активные сессии.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dto.MessageResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     CookieAuth
// @Router       /auth/logout-all [post]
func (c *AuthController) LogoutAll(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if err := c.authSvc.LogoutAll(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.clearTokenCookies(ctx)
	ctx.JSON(http.StatusOK, gin.H{"message": "all sessions terminated"})
}

// ForgotPassword godoc
// @Summary      Запрос сброса пароля
// @Description  Генерирует одноразовый токен сброса (в продакшене отправляется на email, в лабораторной возвращается в ответе).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ForgotPasswordDTO  true  "Email пользователя"
// @Success      200   {object}  dto.ResetTokenResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /auth/forgot-password [post]
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var body dto.ForgotPasswordDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := c.authSvc.ForgotPassword(body.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if token == "" {
		ctx.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link was sent"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"reset_token": token})
}

// ResetPassword godoc
// @Summary      Сброс пароля
// @Description  Устанавливает новый пароль по одноразовому токену. Завершает все активные сессии.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ResetPasswordDTO  true  "Токен и новый пароль"
// @Success      200   {object}  dto.MessageResponse
// @Failure      400   {object}  dto.ErrorResponse    "Неверный или истёкший токен"
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var body dto.ResetPasswordDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.authSvc.ResetPassword(body.Token, body.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}

// OAuthInitiate godoc
// @Summary      Инициация OAuth входа
// @Description  Перенаправляет пользователя на страницу авторизации провайдера (Yandex). Генерирует и сохраняет параметр state для защиты от CSRF.
// @Tags         oauth
// @Param        provider  path  string  true  "Провайдер (yandex)"
// @Success      302  "Редирект на провайдера"
// @Failure      400  {object}  dto.ErrorResponse  "Неподдерживаемый провайдер"
// @Router       /auth/oauth/{provider} [get]
func (c *AuthController) OAuthInitiate(ctx *gin.Context) {
	if ctx.Param("provider") != "yandex" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	state, err := service.GenerateState()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.SetCookie("oauth_state", state, 600, "/", "", false, true)
	ctx.Redirect(http.StatusFound, c.authSvc.GetYandexAuthURL(state))
}

// OAuthCallback godoc
// @Summary      Callback OAuth
// @Description  Обрабатывает ответ от провайдера: проверяет state, обменивает code на токен, находит или создаёт пользователя, устанавливает cookies.
// @Tags         oauth
// @Param        provider  path   string  true  "Провайдер (yandex)"
// @Param        code      query  string  true  "Код авторизации от провайдера"
// @Param        state     query  string  true  "CSRF state параметр"
// @Success      302  "Редирект на главную страницу"
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /auth/oauth/{provider}/callback [get]
func (c *AuthController) OAuthCallback(ctx *gin.Context) {
	if ctx.Param("provider") != "yandex" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	stateCookie, err := ctx.Cookie("oauth_state")
	if err != nil || stateCookie != ctx.Query("state") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	yandexID, email, err := c.authSvc.ExchangeYandexCode(code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "oauth exchange failed"})
		return
	}

	user, err := c.authSvc.FindOrCreateYandexUser(yandexID, email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process user"})
		return
	}

	accessToken, refreshToken, err := c.authSvc.IssueTokenPair(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.setTokenCookies(ctx, accessToken, refreshToken)
	ctx.Redirect(http.StatusFound, "/")
}

func (c *AuthController) setTokenCookies(ctx *gin.Context, accessToken, refreshToken string) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		MaxAge:   int(c.cfg.JWTAccessExpiry.Seconds()),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   int(c.cfg.JWTRefreshExpiry.Seconds()),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *AuthController) clearTokenCookies(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, &http.Cookie{Name: "access_token", Value: "", MaxAge: -1, Path: "/"})
	http.SetCookie(ctx.Writer, &http.Cookie{Name: "refresh_token", Value: "", MaxAge: -1, Path: "/"})
}
