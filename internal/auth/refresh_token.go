package auth

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

const (
	refreshTokenEndpoint          = "https://securetoken.googleapis.com/v1/token"
	errorTokenEmpty        string = "error.google-identity-platform-token-provider.token.empty"
	errorTokenRequestError string = "error.google-identity-platform-token-provider.token.google-request-error"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    string `json:"expiresIn"`
}

type IdentityPlatformRefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
	GrantType    string `json:"grant_type"`
}

type IdentityPlatformRefreshTokenResponse struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
}

func RefreshToken(c *gin.Context) {
	inboundReqBody := RefreshTokenRequest{}
	if err := c.BindJSON(&inboundReqBody); err != nil {
		log.Info().
			Err(err).
			Msg("Error parsing refresh token request body")

		c.JSON(http.StatusBadRequest, reject.BodyParseProblem())
		return
	}

	if strings.TrimSpace(inboundReqBody.RefreshToken) == "" {
		log.Info().
			Msg("Empty refresh token in provider token request")

		c.JSON(http.StatusBadRequest, reject.NewProblem().
			WithTitle("Empty refresh token in provider token request").
			WithStatus(http.StatusBadRequest).
			WithCode(errorTokenEmpty).
			Build())
		return
	}

	outboundReqBody := IdentityPlatformRefreshTokenRequest{
		RefreshToken: inboundReqBody.RefreshToken,
		GrantType:    "refresh_token",
	}

	uri := fmt.Sprintf("%s?key=%s", refreshTokenEndpoint, viper.Get("GOOGLE_PROJECT_API_KEY").(string))
	res, err := http.Post(uri, "application/json", bytes.NewBuffer(utils.JsonEncode(outboundReqBody)))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error calling Google Identity Platform token refresh endpoint")

		c.JSON(http.StatusInternalServerError, reject.NewProblem().
			WithTitle("Failed to exchange refresh token for a new token pair").
			WithStatus(http.StatusInternalServerError).
			WithCode(errorTokenRequestError).
			Build())
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		log.Debug().
			Msgf("Successfully exchanged refresh token for a new Google Identity Platform token pair")

		resBody := utils.JsonDecode[IdentityPlatformRefreshTokenResponse](res.Body)
		c.JSON(http.StatusOK, adaptIdentityPlatformRefreshTokenResponse(resBody))
		return
	}

	errResBody := utils.JsonDecode[GoogleIdentityPlatformErrorResponse](res.Body)

	log.Info().
		Interface("response", errResBody).
		Msg("Failed to exchange refresh token for a new Google Identity Platform token pair")

	problem := reject.NewProblem().
		WithTitle("Failed to exchange refresh token for a new token pair").
		WithStatus(res.StatusCode).
		WithDetail(errResBody.Error.Message).
		WithCode(errorTokenRequestError).
		Build()

	c.JSON(res.StatusCode, problem)
}

func adaptIdentityPlatformRefreshTokenResponse(response IdentityPlatformRefreshTokenResponse) RefreshTokenResponse {
	rt := RefreshTokenResponse{}
	rt.IDToken = response.IDToken
	rt.TokenType = response.TokenType
	rt.RefreshToken = response.RefreshToken
	rt.ExpiresIn = response.ExpiresIn
	return rt
}
