package profile

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type profileHandler struct {
	profile *ProfileService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := profileHandler{
		profile: &ProfileService{Db: db},
	}

	routes := rg.Group("/profile")
	routes.GET("/:id", middleware.VerifyAuthToken, handler.getProfileById)
	routes.PUT("/:id/blocks", middleware.VerifyAuthToken, handler.activateBlock)
}

func (h profileHandler) getProfileById(c *gin.Context) {
	id, parseErr := strconv.ParseUint(c.Param("id"), 0, 64)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, reject.RequestParamsProblem())
		return
	}

	profile, err := h.profile.FindById(id)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, profile)
}

type ActivateBlocksRequest struct {
	activeBlockIds []uint64
}

func (h profileHandler) activateBlock(c *gin.Context) {
	userId, parseErr := strconv.ParseUint(c.Param("id"), 0, 64)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, reject.RequestParamsProblem())
		return
	}

	body := ActivateBlocksRequest{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	err := h.profile.activateBlocks(userId, body)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.Status(http.StatusNoContent)
}
