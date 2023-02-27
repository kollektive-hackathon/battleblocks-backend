package game

import (
	"context"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"time"

	gcppubsub "cloud.google.com/go/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Moved struct {
	GameId        uint64 `json:"gameID"`
	PlayerId      uint64 `json:"gamePlayerID"`
	PlayerAddress uint64 `json:"playerAddress"`
	X             uint   `json:"coordinateX"`
	Y             uint   `json:"coordinateY"`
}

type GameOver struct {
	GameId         uint64          `json:"gameID"`
	PlayerA        uint64          `json:"playerA"`
	PlayerB        uint64          `json:"playerB"`
	Winner         uint64          `json:"winner"`
	PlayerHitCount map[string]uint `json:"PlayerHitCount"`
}

type GameCreated struct {
	GameId         uint64 `json:"gameId"`
	CreatorId      uint64 `json:"creatorId"`
	CreatorAddress string `json:"creatorAddress"`
	Stake          uint64 `json:"stake"`
}

type gameContractBridge struct {
	db *gorm.DB
}

func (b *gameContractBridge) sendCreateGameTx(stake float32, rootMerkel string, userAuthorizer blockchain.Authorizer) {
	commandType := "GAME_CREATE"
	payload := []any{
		stake,
		rootMerkel,
	}
	authorizers := []blockchain.Authorizer{userAuthorizer, blockchain.GetAdminAuthorizer()}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (b *gameContractBridge) sendMove(
	gameId uint64,
	guessX uint64,
	guessY uint64,
	proof *[][]uint8,
	blockPresent *bool,
	opponentGuessX *uint64,
	opponentGuessY *uint64,
	nonce *uint64,
) {
	commandType := "GAME_MOVE"
	payload := []any{
		stake,
		rootMerkel,
	}
	authorizers := []blockchain.Authorizer{userAuthorizer, blockchain.GetAdminAuthorizer()}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (b *gameContractBridge) handleMoved(_ context.Context, message *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(message.Data))
	messagePayload, err := utils.JsonDecodeByteStream[Moved](message.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Error while parsing Moved message")
		return
	}

	mh := model.MoveHistory{
		UserId:      messagePayload.PlayerId,
		GameId:      messagePayload.GameId,
		CoordinateX: messagePayload.X,
		CoordinateY: messagePayload.Y,
		PlayedAt:    time.Now().UTC(),
	}

	result := b.db.Create(&mh)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling Moved")
		return
	}

	message.Ack()
}

func (b *gameContractBridge) handleGameCreated(_ context.Context, message *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(message.Data))
	messagePayload, err := utils.JsonDecodeByteStream[GameCreated](message.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Error while parsing GameOver message")
		return
	}

	timeNow := time.Now().UTC()
	game := model.Game{
		OwnerId:     messagePayload.CreatorId,
		GameStatus:  "CREATED",
		Stake:       messagePayload.Stake,
		TimeCreated: timeNow,
	}

	result := b.db.Create(&game)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling GameOver")
		return
	}

	message.Ack()
}

func (b *gameContractBridge) handleGameOver(_ context.Context, message *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(message.Data))
	messagePayload, err := utils.JsonDecodeByteStream[GameOver](message.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Error while parsing GameOver message")
		return
	}

	result := b.db.
		Model(&model.Game{}).
		Where("id = ?", messagePayload.GameId).
		Updates(map[string]any{
			"winner":      messagePayload.Winner,
			"game_status": "FINISHED",
		})

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling GameOver")
		return
	}

	message.Ack()
}
