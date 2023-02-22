package ws

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/ws"
	"github.com/rs/zerolog/log"
)

type wsHandler struct {
	notificationHub *ws.WebSocketNotificationHub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func RegisterRoutes(rg *gin.RouterGroup) {
	handler := wsHandler{
		notificationHub: ws.NewNotificationHub(),
	}

	routes := rg.Group("/ws")
	routes.GET("/game/:id", middleware.VerifyAuthToken, handler.serveWs)
}

func (wsh *wsHandler) serveWs(c *gin.Context) {
	gameId := c.Param("gameId")
	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	defer wsh.notificationHub.UnregisterListener(gameId, conn)

	wsh.notificationHub.RegisterListener(gameId, conn)

	for {
		var buffer any
		err := conn.ReadJSON(&buffer)
		if err != nil {
			log.Warn().Err(err).Msg("Error reading ws message")
			return
		}
	}
}
