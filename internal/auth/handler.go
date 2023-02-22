package auth

import (
	"github.com/gin-gonic/gin"
)

const tokenEndpoint = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithIdp"

type authHandler struct {
}

func RegisterRoutes(rg *gin.RouterGroup) {
	handler := &authHandler{}

	routes := rg.Group("/auth")
	routes.POST("/google", handler.getIdentityPlatformTokenFromGoogleIdToken)
	routes.POST("/apple", handler.getIdentityPlatformTokenFromAppleIdToken)
	routes.POST("/refresh", RefreshToken)
}
