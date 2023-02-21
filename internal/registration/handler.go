package registration

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type registrationHandler struct {
	db *gorm.DB
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {

}
