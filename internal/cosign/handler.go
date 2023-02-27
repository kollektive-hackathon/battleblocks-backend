package cosign

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"gorm.io/gorm"
	"net/http"
)

type cosignHandler struct {
	cosign *cosignService
}

type CosignRequest struct {
	Method  string         `json:"method"`
	Payload map[string]any `json:"payload"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := &cosignHandler{
		cosign: &cosignService{db: db},
	}

	routes := rg.Group("/cosign")
	routes.POST("", middleware.VerifyAuthToken, handler.handleCosign)
}

func (ch cosignHandler) handleCosign(c *gin.Context) {
	body := CosignRequest{}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	signedVoucher, err := ch.cosign.VerifyAndSign(utils.GetAccessToken(c), body)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, signedVoucher)
}
