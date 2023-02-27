package game

import (
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
	gameService gameService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := gameHandler{
		gameService: gameService{
			db: db,
			gameContractBridge: &gameContractBridge{
				db: db,
			},
		},
	}

	routes := rg.Group("/game")
	// routes.GET("/", handler.getGames)
	routes.POST("", middleware.VerifyAuthToken, handler.createGame)
	routes.GET("", middleware.VerifyAuthToken, handler.getGames)

	routes.GET("/:id/moves", middleware.VerifyAuthToken, handler.getMoves)
	routes.POST("/:id/moves", middleware.VerifyAuthToken, handler.playMove)
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
	X uint   `json:"x"`
	Y uint64 `json:"y"`
}

func (gh *gameHandler) playMove(c *gin.Context) {
	body := PlayMoveRequest{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	userEmail := utils.GetUserEmail(c)

	// TODO dobijem coordinate za shoot

	// TODO iz move history dohvati od protivnika zadnji move i napravi proof
	//    - iz block placement mojih kreirati merkle tree i pomocu nonce tablice dohvatit
	//   sve (isPresent, x, y, nonce) za svog usera i taj game (fabsa savea nonceve prilikom game create)

	// TODO iz kreiranog merkle treea dobit proof za field koji gada trenutno

	// TODO posalji command on chain
}

// JOIN Game
// JOIN game - gameID i PLACEMENTS -- create merkel -- tx join - CHECK BALANCE

// MOVE endpoint
//  kordinate za shoot se dobiju
//  Poslat TX --> Koordinate, PROOF, ORIGINAL LIST
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
	//placements also sent here
	//TODO: CHECK balance ---- execute script  with golang-flow-sdk
	//
	// create merkeltree, save it to DB under GAME.owner_root_merkle, send tx - with the ROOT and stake.

	body := CreateGameRequest{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}
	userId := utils.GetUserId(c)

	err := gh.gameService.createGame(body, userId)

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
