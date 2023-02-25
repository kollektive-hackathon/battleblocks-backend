package auth

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	reject2 "github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/profile"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

type IDTokenRequest struct {
	IDToken     string `json:"idToken"`
	AccessToken string `json:"accessToken"`
}

type IdentityPlatformTokenRequest struct {
	PostBody            string `json:"postBody"`
	RequestURI          string `json:"requestUri"`
	ReturnIDPCredential bool   `json:"returnIdpCredential"`
	ReturnSecureToken   bool   `json:"returnSecureToken"`
}

type IdentityPlatformTokenResponse struct {
	Email         string           `json:"email"`
	EmailVerified bool             `json:"emailVerified"`
	LocalID       string           `json:"localId"`
	IDToken       string           `json:"idToken"`
	RefreshToken  string           `json:"refreshToken"`
	ExpiresIn     string           `json:"expiresIn"`
	Profile       *profile.Profile `json:"profile"`
}

func (ah authHandler) getIdentityPlatformTokenFromGoogleIdToken(c *gin.Context) {
	ah.getIdentityPlatformTokenFromProviderIDToken(c, "google.com")
}

func (ah authHandler) getIdentityPlatformTokenFromAppleIdToken(c *gin.Context) {
	ah.getIdentityPlatformTokenFromProviderIDToken(c, "apple.com")
}

func (ah authHandler) getIdentityPlatformTokenFromProviderIDToken(c *gin.Context, provider string) {
	inboundReqBody := IDTokenRequest{}
	if err := c.BindJSON(&inboundReqBody); err != nil {
		log.Info().
			Err(err).
			Msg("Error parsing ID token request body")

		c.JSON(http.StatusBadRequest, reject2.BodyParseProblem())
		return
	}

	isIDTokenEmpty := strings.TrimSpace(inboundReqBody.IDToken) == ""
	isAccessTokenEmpty := strings.TrimSpace(inboundReqBody.AccessToken) == ""

	if isIDTokenEmpty && isAccessTokenEmpty {
		log.Info().
			Msg("Empty ID and access token in provider token request")

		c.JSON(http.StatusBadRequest, reject2.NewProblem().
			WithTitle("Either idToken or accessToken must be passed").
			WithStatus(http.StatusBadRequest).
			WithCode(errorTokenEmpty).
			Build())
		return
	}

	postBody := fmt.Sprintf("providerId=%s", provider)
	if isIDTokenEmpty {
		postBody = fmt.Sprintf("%s&access_token=%s", postBody, inboundReqBody.AccessToken)
	} else {
		postBody = fmt.Sprintf("%s&id_token=%s", postBody, inboundReqBody.IDToken)
	}

	outboundReqBody := IdentityPlatformTokenRequest{
		PostBody:            postBody,
		RequestURI:          "http://internal", // arbitrary requestUri since it makes no difference in our case
		ReturnIDPCredential: true,
		ReturnSecureToken:   true,
	}

	uri := fmt.Sprintf("%s?key=%s", tokenEndpoint, viper.Get("GOOGLE_PROJECT_API_KEY").(string))
	res, err := http.Post(uri, "application/json", bytes.NewBuffer(utils.JsonEncode(outboundReqBody)))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error calling Google Identity Platform idp keymgmt in endpoint")

		c.JSON(http.StatusInternalServerError, reject2.NewProblem().
			WithTitle("Failed to exchange provider ID token for an internal token pair").
			WithStatus(http.StatusInternalServerError).
			WithCode(errorTokenRequestError).
			Build())
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		log.Debug().
			Msgf("Successfully exchanged %s ID token for a Google Identity Platform token pair", provider)

		resBody := utils.JsonDecode[IdentityPlatformTokenResponse](res.Body)

		var p profile.Profile
		result := ah.db.
			Table("user").
			Joins("INNER JOIN custodial_wallet ON user.custodial_wallet_id = custodial_wallet.id").
			Where("user.email = ?", resBody.Email).
			Select(`
			user.id, 
			user.email,
			user.username,
			custodial_wallet.address AS custodial_wallet_address,
			user.self_custody_wallet_address AS self_custody_wallet_address
		`).Scan(&p)

		if result.Error == nil {
			resBody.Profile = &p
		}

		c.JSON(http.StatusOK, resBody)
		return
	}

	errResBody := utils.JsonDecode[GoogleIdentityPlatformErrorResponse](res.Body)

	log.Info().
		Interface("response", errResBody).
		Msgf("Failed to exchange %s ID token for a Google Identity Platform token pair", provider)

	problem := reject2.NewProblem().
		WithTitle("Failed to exchange provider ID token for an internal token pair").
		WithStatus(res.StatusCode).
		WithDetail(errResBody.Error.Message).
		WithCode(errorTokenRequestError).
		Build()

	c.JSON(res.StatusCode, problem)
}
