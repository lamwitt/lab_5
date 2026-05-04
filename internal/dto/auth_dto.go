package dto

import (
	"time"

	"github.com/google/uuid"
)

type RegisterDTO struct {
	Email    string `json:"email"    binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

type LoginDTO struct {
	Email    string `json:"email"    binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required"       example:"password123"`
}

type ForgotPasswordDTO struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type ResetPasswordDTO struct {
	Token    string `json:"token"    binding:"required" example:"a3f1c2d4..."`
	Password string `json:"password" binding:"required,min=8" example:"newpassword123"`
}

type UserResponseDTO struct {
	ID        uuid.UUID `json:"id"        example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email"     example:"user@example.com"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T10:00:00Z"`
}

type MessageResponse struct {
	Message string `json:"message" example:"success"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error description"`
}

type ResetTokenResponse struct {
	ResetToken string `json:"reset_token" example:"a3f1c2d4e5b6..."`
}
