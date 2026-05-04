package middleware

import (
	"net/http"

	"books-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthRequired(authSvc *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("access_token")
		if err != nil || token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := authSvc.ValidateAccessToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("accessToken", token)
		ctx.Next()
	}
}

func GetUserID(ctx *gin.Context) uuid.UUID {
	return ctx.MustGet("userID").(uuid.UUID)
}
