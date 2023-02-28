package game

import (
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/ws"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"gorm.io/gorm"
)

type gameHandler struct {
	gameService *gameService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := gameHandler{
		gameService: &gameService{
			db: db,
			gameContractBridge: &gameContractBridge{
				db:              db,
				notificationHub: ws.NewNotificationHub(),
			},
		},
	}

	routes := rg.Group("/game")
	routes.GET("", middleware.VerifyAuthToken, handler.getGames)
	routes.GET("/:id", middleware.VerifyAuthToken, handler.getGame)
	routes.POST("", middleware.VerifyAuthToken, handler.createGame)
	routes.POST("/:id/join", middleware.VerifyAuthToken, handler.joinGame)

	routes.GET("/:id/moves", middleware.VerifyAuthToken, handler.getMoves)
	routes.POST("/:id/moves", middleware.VerifyAuthToken, handler.playMove)

	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.move-done-sub",
		Handler:        handler.gameService.gameContractBridge.handleMoved,
	})

	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.game-created-sub",
		Handler:        handler.gameService.gameContractBridge.handleGameCreated,
	})

	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.challenger-joined-sub",
		Handler:        handler.gameService.gameContractBridge.handleChallengerJoined,
	})

	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.game-over-sub",
		Handler:        handler.gameService.gameContractBridge.handleGameOver,
	})
}

func (gh *gameHandler) getMoves(c *gin.Context) {
	gameId, parseErr := strconv.ParseUint(c.Param("id"), 0, 64)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, reject.RequestParamsProblem())
		return
	}

	moves, err := gh.gameService.getMoves(gameId, utils.GetUserEmail(c))
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, moves)
}

type PlayMoveRequest struct {
	X uint64 `json:"x"`
	Y uint64 `json:"y"`
}

func (gh *gameHandler) playMove(c *gin.Context) {
	gameId, parseErr := strconv.ParseUint(c.Param("id"), 0, 64)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, reject.RequestParamsProblem())
		return
	}

	body := PlayMoveRequest{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	userEmail := utils.GetUserEmail(c)

	gh.gameService.playMove(gameId, userEmail, body)

	c.Status(http.StatusNoContent)
}

func (gh *gameHandler) getGame(c *gin.Context) {
	gameId, parseErr := strconv.ParseUint(c.Param("id"), 0, 64)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, reject.RequestParamsProblem())
		return
	}
	game, err := gh.gameService.getGame(gameId)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, game)
}
func (gh *gameHandler) getGames(c *gin.Context) {
	page, err := utils.NewPageRequest(c)
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	userEmail := utils.GetUserEmail(c)
	games, gamesCount, err := gh.gameService.getGames(page, userEmail)
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
	body := CreateGameRequest{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	userEmail := utils.GetUserEmail(c)
	createdGame, err := gh.gameService.createGame(body, userEmail)

	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, createdGame)
}

func (gh *gameHandler) joinGame(c *gin.Context) {
	body := JoinGameRequest{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}
	gameId, parseErr := strconv.ParseUint(c.Param("id"), 0, 64)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, reject.RequestParamsProblem())
		return
	}

	userEmail := utils.GetUserEmail(c)
	err := gh.gameService.joinGame(body, gameId, userEmail)

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
