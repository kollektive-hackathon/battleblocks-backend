package main

import (
	"github.com/kollektive-hackathon/battleblocks-backend/internal/paypal"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/auth"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/cosign"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/game"
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
)

func main() {
	setupViper()
	setupZerolog()
	pubsub.InitPubSub()
	db := setupDb()
	apiRouter := setupApiRouter(db)

	defer func() { pubsub.CloseClient() }()

	firebase.InitFirebaseSdk()

	server := &http.Server{
		Addr:         ":8000",
		Handler:      apiRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Warn().Interface("err", err.Error()).Msg("Could not start server")
	}
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

	// gcp health check
	apiRouter.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	middleware.RegisterGlobalMiddleware(apiRouter)
	routerGroup := apiRouter.Group("/api")

	ws.RegisterRoutes(routerGroup)
	auth.RegisterRoutes(routerGroup, db)
	paypal.RegisterRoutes(routerGroup)
	registration.RegisterRoutesAndSubscriptions(routerGroup, db)
	profile.RegisterRoutes(routerGroup, db)
	shop.RegisterRoutesAndSubscriptions(routerGroup, db)
	game.RegisterRoutes(routerGroup, db)
	cosign.RegisterRoutes(routerGroup, db)

	return apiRouter
}

func setupViper() {
	viper.AutomaticEnv()
	viper.SetConfigFile("./.env")
	viper.ReadInConfig()
}

func setupZerolog() {
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "time"
	zerolog.TimeFieldFormat = time.RFC3339Nano
}
