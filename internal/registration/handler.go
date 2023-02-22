package registration

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/reject"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type registrationHandler struct {
	registration registrationService
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	handler := registrationHandler{
		registration: registrationService{db: db},
	}

	routes := rg.Group("/registration")
	routes.POST("/", middleware.VerifyAuthToken, handler.register)
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

	// TODO get token claims - email? ...

	h.registration.register(username)
}
