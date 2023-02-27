package paypal

import (
	"encoding/json"
	"fmt"
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
	// description -- user id
	// custom_id -- item id
	userId := body["description"]
	blockId := body["custom_id"]
	fmt.Printf("%v, %v", userId, blockId)
}
