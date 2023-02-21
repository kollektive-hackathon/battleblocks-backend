package utils

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	emailClaimKey string = "email"
	tokenCtxKey   string = "accessToken"
	userIdCtxKey  string = "userId"
)

type AccessToken struct {
	Token    auth.Token
	RawToken string
}

func GetAccessToken(ctx *gin.Context) auth.Token {
	at := getAccessToken(ctx)
	return at.Token
}

func GetAccessTokenRaw(ctx *gin.Context) string {
	at := getAccessToken(ctx)
	return at.RawToken
}

func getAccessToken(ctx *gin.Context) AccessToken {
	return getCtxValue(tokenCtxKey, ctx).(AccessToken)
}

func GetUserEmail(ctx *gin.Context) string {
	token := GetAccessToken(ctx)
	return token.Claims[emailClaimKey].(string)
}

func GetUserExternalId(ctx *gin.Context) string {
	token := GetAccessToken(ctx)
	return token.Subject
}

func getCtxValue(key string, ctx *gin.Context) any {
	value, exists := ctx.Get(key)
	if !exists {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
	return value
}

func SetAccessTokenCtx(token *AccessToken, ctx *gin.Context) {
	ctx.Set(tokenCtxKey, *token)
}

func SetUserIdCtx(userId uint64, ctx *gin.Context) {
	ctx.Set(userIdCtxKey, userId)
}
