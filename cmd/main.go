// @title           Books API
// @version         1.0
// @description     REST API для управления каталогом книг с JWT-аутентификацией и OAuth 2.0 (Yandex).
// @host            localhost:4200
// @BasePath        /

// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        access_token
// @description                 HttpOnly cookie с access-токеном. Установите через POST /auth/login, затем браузер будет отправлять его автоматически.

package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"

	_ "books-api/docs"
	"books-api/internal/cache"
	"books-api/internal/config"
	"books-api/internal/controller"
	"books-api/internal/database"
	"books-api/internal/repository"
	"books-api/internal/service"
	"books-api/migrations"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.RunMigrations(cfg, migrations.FS); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	cacheSvc := cache.NewCacheService(cfg)

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	bookRepo := repository.NewBookRepository(db)

	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg, cacheSvc)
	bookSvc := service.NewBookService(bookRepo, cacheSvc, cfg)

	authCtrl := controller.NewAuthController(authSvc, cfg)
	bookCtrl := controller.NewBookController(bookSvc, authSvc)

	r := gin.Default()
	controller.RegisterFrontendRoutes(r)
	authCtrl.RegisterRoutes(r)
	bookCtrl.RegisterRoutes(r)

	if cfg.SwaggerEnabled {
		r.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Println("Swagger UI available at http://localhost:" + cfg.Port + "/api/docs/index.html")
	}

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
