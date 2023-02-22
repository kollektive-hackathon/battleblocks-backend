package shop

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"gorm.io/gorm"
	"net/http"
)

type shopHandler struct {
	shop shopService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := shopHandler{
		shop: shopService{db: db},
	}

	routes := rg.Group("/shop")
	routes.GET("/", middleware.VerifyAuthToken, handler.getShopList)
}

func (h shopHandler) getShopList(c *gin.Context) {
	blocks, err := h.shop.FindAll()
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, blocks)
}
