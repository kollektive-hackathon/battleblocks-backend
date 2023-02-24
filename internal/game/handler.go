package game

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"gorm.io/gorm"
)

type gameHandler struct {
	gameService gameService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := gameHandler{
		gameService: gameService{db: db},
	}

	routes := rg.Group("/game")
	// routes.GET("/", handler.getGames)
	routes.POST("/", middleware.VerifyAuthToken, handler.createGame)
	routes.GET("/", middleware.VerifyAuthToken, handler.getGames)
}

func (gh *gameHandler) getGames(c *gin.Context) {
	page, err := utils.NewPageRequest(c)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	userId := utils.GetUserId(c)
	games, gamesCount, err := gh.gameService.getGames(page, userId)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	response := utils.NewPageResponse[model.Game]().
		WithItems(games).
		WithItemCount(*gamesCount)

	nextToken := checkNextPageToken(page, *gamesCount)
	if nextToken != nil {
		response.WithNextPageToken(*nextToken)
	}

	c.JSON(http.StatusOK, response.Build())
}

func (gh *gameHandler) createGame(c *gin.Context) {
	body := model.Game{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	err := gh.gameService.createGame(body)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
	}
}

func checkNextPageToken(currPage utils.PageRequest, gameCount int64) *int64 {
	if int(gameCount) > (currPage.Token+1)*currPage.Size {
		nextToken := int64(currPage.Token + 1)
		return &nextToken
	}
	return nil
}
