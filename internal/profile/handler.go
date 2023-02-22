package profile

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"gorm.io/gorm"
)

type profileHandler struct {
	profile profileService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := profileHandler{
		profile: profileService{db: db},
	}

	routes := rg.Group("/profile")
	routes.POST("/", middleware.VerifyAuthToken, handler.getProfile)
}

func (h profileHandler) getProfile(context *gin.Context) {

}
