package registration

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/ws"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/profile"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type registrationHandler struct {
	registration *registrationService
	profile      *profile.ProfileService
}

func RegisterRoutesAndSubscriptions(rg *gin.RouterGroup, db *gorm.DB) {
	handler := registrationHandler{
		registration: &registrationService{
			db: db,
			bridge: &accountContractBridge{
				db:              db,
				profileService:  &profile.ProfileService{Db: db},
				notificationHub: ws.NewNotificationHub(),
			},
		},
		profile: &profile.ProfileService{Db: db},
	}

	routes := rg.Group("/registration")
	routes.POST("", middleware.VerifyAuthToken, handler.register)

	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.account-created-sub",
		Handler:        handler.registration.bridge.handleCustodialAccountCreated,
	})

	go pubsub.Subscribe(pubsub.SubscriptionHandler{
		SubscriptionId: "blockchain.flow.events.account-delegated-sub",
		Handler:        handler.registration.bridge.handleCustodialAccountDelegated,
	})
}

type RegistrationRequest struct {
	Username string `json:"username"`
}

func (h registrationHandler) register(c *gin.Context) {
	body := RegistrationRequest{}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" || len(username) > 32 {
		c.JSON(http.StatusBadRequest, reject.RequestValidationProblem())
		return
	}

	userId, err := h.registration.register(username, utils.GetUserEmail(c), utils.GetUserExternalId(c))

	if err != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	createdProfile, profileLoadErr := h.profile.FindById(userId)
	if profileLoadErr != nil {
		c.JSON(err.Problem.Status, err.Problem)
		return
	}

	c.JSON(http.StatusOK, createdProfile)
}
