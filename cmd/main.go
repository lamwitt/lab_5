package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

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

	bookRepo := repository.NewBookRepository(db)
	bookSvc := service.NewBookService(bookRepo)
	bookCtrl := controller.NewBookController(bookSvc)

	r := gin.Default()
	bookCtrl.RegisterRoutes(r)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
