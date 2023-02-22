package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/auth"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/firebase"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/middleware"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/profile"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/registration"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/shop"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/ws"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func main() {
	setupViper()
	setupZerolog()
	db := setupDb()
	apiRouter := setupApiRouter(db)

	pubsub.InitPubSub()
	defer func() { pubsub.CloseClient() }()

	firebase.InitFirebaseSdk()

	port := viper.Get("PORT").(string)
	server := &http.Server{
		Addr:         port,
		Handler:      apiRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server.ListenAndServe()
}

func setupDb() *gorm.DB {
	dbUrl := viper.Get("DB_URL").(string)

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	sqlDb, _ := db.DB()

	sqlDb.SetMaxOpenConns(50)
	sqlDb.SetConnMaxLifetime(time.Minute * 10)

	return db
}

func setupApiRouter(db *gorm.DB) *gin.Engine {
	apiRouter := gin.Default()
	routerGroup := apiRouter.Group("/battleblocks-api")

	middleware.RegisterGlobalMiddleware(apiRouter)

	ws.RegisterRoutes(routerGroup)
	auth.RegisterRoutes(routerGroup)
	registration.RegisterRoutes(routerGroup, db)
	profile.RegisterRoutes(routerGroup, db)
	shop.RegisterRoutes(routerGroup, db)

	return apiRouter
}

func setupViper() {
	viper.AutomaticEnv()

	// kms
	viper.SetDefault("GOOGLE_KMS_PROJECT_ID", "0")
	viper.SetDefault("GOOGLE_KMS_LOCATION_ID", "0")
	viper.SetDefault("GOOGLE_KMS_KEYRING_ID", "0")
}

func setupZerolog() {
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "time"
	zerolog.TimeFieldFormat = time.RFC3339Nano
}
