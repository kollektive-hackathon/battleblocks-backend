package ws

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/ws"
	"github.com/rs/zerolog/log"
)

type wsHandler struct {
	notificationHub *ws.WebSocketNotificationHub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RegisterRoutes(rg *gin.RouterGroup) {
	handler := wsHandler{
		notificationHub: ws.NewNotificationHub(),
	}

	routes := rg.Group("/ws")
	routes.GET("/game/:id", handler.serveGameWs)
	routes.GET("/registration/:userEmail", handler.serveRegistrationWs)
}

func (wsh *wsHandler) serveGameWs(c *gin.Context) {
	gameId := c.Param("id")
	conn, er := upgrader.Upgrade(c.Writer, c.Request, nil)
	if er != nil {
		log.Warn().Err(er).Msg("Couldnt upgrade request")
		return
	}

	defer wsh.notificationHub.UnregisterListener(fmt.Sprintf("game/%s", gameId), conn)

	wsh.notificationHub.RegisterListener(fmt.Sprintf("game/%s", gameId), conn)

	for {
		var buffer any
		err := conn.ReadJSON(&buffer)
		if err != nil {
			log.Warn().Err(err).Msg("Error reading ws message")
			return
		}
	}
}

func (wsh *wsHandler) serveRegistrationWs(c *gin.Context) {
	//userEmail := utils.GetUserEmail(c)
	userEmail := c.Param("userEmail")
	conn, er := upgrader.Upgrade(c.Writer, c.Request, nil)
	if er != nil {
		log.Warn().Err(er).Msg("Couldnt upgrade request")
		return
	}
	defer wsh.notificationHub.UnregisterListener(fmt.Sprintf("registration/%s", userEmail), conn)

	wsh.notificationHub.RegisterListener(fmt.Sprintf("registration/%s", userEmail), conn)

	for {
		var buffer any
		err := conn.ReadJSON(&buffer)
		if err != nil {
			log.Warn().Err(err).Msg("Error reading ws message")
			return
		}
	}
}
