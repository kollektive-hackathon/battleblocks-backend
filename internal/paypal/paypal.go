package paypal

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/pp")
	routes.POST("/hook", handlePaypalWebhook)
}

func handlePaypalWebhook(c *gin.Context) {
	rawBody, _ := ioutil.ReadAll(c.Request.Body)
	var body map[string]any

	json.Unmarshal(rawBody, &body)

}
