package shop

import (
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"net/http"

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
	routes.GET("/", middleware.VerifyAuthToken, handler.getShopList)

	// TODO subscription ids
	// pubsub.Subscribe(pubsub.SubscriptionHandler{
	// SubscriptionId: "",
	// Handler:        handler.shop.bridge.handleWithdrew,
	// })
	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.minted",
		Handler:        handler.shop.bridge.handleMinted,
	})
	// pubsub.Subscribe(pubsub.SubscriptionHandler{
	// SubscriptionId: "",
	// Handler:        handler.shop.bridge.handleDeposited,
	// })
	// pubsub.Subscribe(pubsub.SubscriptionHandler{
	// SubscriptionId: "",
	// Handler:        handler.shop.bridge.handleBurned,
	// })
}

func (h shopHandler) getShopList(c *gin.Context) {
	blocks, err := h.shop.FindAll()
	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, blocks)
}
