package middleware

import (
	"github.com/gin-gonic/gin"
)

func RegisterGlobalMiddleware(router *gin.Engine) {
	router.Use(gin.Recovery(), CORS())
}
