package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/profile"
	"gorm.io/gorm"
)

const tokenEndpoint = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithIdp"

type authHandler struct {
	db      *gorm.DB
	profile *profile.ProfileService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := &authHandler{db: db, profile: &profile.ProfileService{Db: db}}

	routes := rg.Group("/auth")
	routes.POST("/google", handler.getIdentityPlatformTokenFromGoogleIdToken)
	routes.POST("/apple", handler.getIdentityPlatformTokenFromAppleIdToken)
	routes.POST("/refresh", RefreshToken)
}
