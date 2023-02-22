package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/firebase"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/utils"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

const (
	accessTokenRequired string = "error.token.required"
	accessTokenInvalid  string = "error.token.invalid"
)

func VerifyAuthToken(context *gin.Context) {
	authHeader := context.Request.Header.Get("Authorization")
	idTokenValue := strings.TrimSpace(strings.ReplaceAll(authHeader, "Bearer", ""))
	if idTokenValue == "" {
		log.Warn().Msg("Token missing: 401")
		context.AbortWithStatusJSON(
			http.StatusUnauthorized,
			reject.NewProblem().
				WithTitle("Missing access token").
				WithStatus(http.StatusUnauthorized).
				WithCode(accessTokenRequired).
				Build())
		return
	}
	token, err := firebase.VerifyIdToken(idTokenValue)
	if err != nil {
		log.Warn().Msg(fmt.Sprintf("Error verifying token: %s", err.Error()))
		context.AbortWithStatusJSON(
			http.StatusUnauthorized,
			reject.NewProblem().
				WithTitle("Cannot verify access token").
				WithStatus(http.StatusUnauthorized).
				WithCode(accessTokenInvalid).
				WithDetail(err.Error()).
				Build())
		return
	}
	accessTokenDetails := utils.AccessToken{
		Token:    *token,
		RawToken: idTokenValue,
	}
	utils.SetAccessTokenCtx(&accessTokenDetails, context)
}
