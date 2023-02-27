package profile

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"gorm.io/gorm"
	"net/http"
)

type profileHandler struct {
	profile *ProfileService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := profileHandler{
		profile: &ProfileService{Db: db},
	}

	routes := rg.Group("/profile")
	routes.GET("", middleware.VerifyAuthToken, handler.getProfile)
	routes.PUT("/blocks", middleware.VerifyAuthToken, handler.activateBlock)
}

func (h profileHandler) getProfile(c *gin.Context) {
	email := utils.GetUserEmail(c)

	profile, err := h.profile.FindByEmail(email)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, profile)
}

type ActivateBlocksRequest struct {
	ActiveBlockIds []uint64 `json:"activeBlockIds"`
}

func (h profileHandler) activateBlock(c *gin.Context) {
	userEmail := utils.GetUserEmail(c)

	body := ActivateBlocksRequest{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	err := h.profile.activateBlocks(userEmail, body)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	profile, err := h.profile.FindByEmail(userEmail)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, profile)
}
