package firebase

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/rs/zerolog/log"
)

var firebaseAuthClient *auth.Client
var ctx context.Context

func InitFirebaseSdk() {
	ctx = context.Background()
	app, appErr := firebase.NewApp(ctx, nil)
	if appErr != nil {

		log.Fatal().Err(appErr).Msg("error initializing app")
	}
	var clientErr error
	firebaseAuthClient, clientErr = app.Auth(ctx)
	if clientErr != nil {
		log.Fatal().Err(clientErr).Msg("error getting Auth client")
	}
}

func VerifyIdToken(idToken string) (*auth.Token, error) {
	return firebaseAuthClient.VerifyIDToken(ctx, idToken)
}
