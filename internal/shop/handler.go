package shop

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"gorm.io/gorm"
)

type shopHandler struct {
	shop shopService
}

func RegisterRoutesAndSubscriptions(rg *gin.RouterGroup, db *gorm.DB) {
	handler := shopHandler{
		shop: shopService{
			db: db,
			bridge: &nftContractBridge{
				db: db,
			},
		},
	}

	routes := rg.Group("/shop")
	routes.GET("", middleware.VerifyAuthToken, handler.getShopList)
	routes.POST("/pp", handler.paypalWebhook)

	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.minted",
		Handler:        handler.shop.bridge.handleMinted,
	})
}

func (h shopHandler) getShopList(c *gin.Context) {
	blocks, err := h.shop.FindAll()
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, blocks)
}

func (h shopHandler) paypalWebhook(c *gin.Context) {
	rawBody, _ := ioutil.ReadAll(c.Request.Body)
	var body map[string]any
	json.Unmarshal(rawBody, &body)
	log.Info().Interface("pp_data", body).Msg("Pp data--")

	// description -- user id
	// custom_id -- item id
	userId := body["description"].(string)
	blockId := body["custom_id"].(string)

	h.shop.SendBoughtToUser(userId, blockId)
}
