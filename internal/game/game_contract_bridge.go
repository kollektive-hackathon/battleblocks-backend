package game

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/ws"

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

type ChallengerJoined struct {
	GameId            uint64           `json:"gameId"`
	StartTime         uint64           `json:"startTime"`
	Wager             float64          `json:"wager"`
	PlayerA           string           `json:"playerA"`
	PlayerB           string           `json:"playerB"`
	Winner            string           `json:"winner"`
	PlayerHitCount    map[string]uint8 `json:"playerHitCount"`
	GameState         uint8            `json:"gameState"`
	Turn              uint8            `json:"turn"`
	PlayerAMerkleRoot []uint8          `json:"playerAMerkleRoot"`
	PlayerBMerkleRoot []uint8          `json:"playerBMerkleRoot"`
}

type GameOver struct {
	GameId         uint64          `json:"gameID"`
	PlayerA        uint64          `json:"playerA"`
	PlayerB        uint64          `json:"playerB"`
	Winner         uint64          `json:"winner"`
	PlayerHitCount map[string]uint `json:"playerHitCount"`
}

type GameCreated struct {
	GameId         uint64 `json:"gameID"`
	CreatorId      uint64 `json:"creatorID"`
	CreatorAddress string `json:"creatorAddress"`
	Stake          uint64 `json:"wager"`
	Payload        uint64 `json:"payload"`
}

type gameContractBridge struct {
	db              *gorm.DB
	notificationHub *ws.WebSocketNotificationHub
}

func (b *gameContractBridge) sendJoinGame(stake float32, rootMerkel []byte, gameId uint64, userAuthorizer blockchain.Authorizer) {
	commandType := "GAME_JOIN"
	uint8Merkle := byteArrayToUint(rootMerkel)
	payload := []any{
		stake,
		uint8Merkle,
		gameId,
	}
	authorizers := []blockchain.Authorizer{userAuthorizer, blockchain.GetAdminAuthorizer()}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (b *gameContractBridge) sendCreateGameTx(stake float32, rootMerkel []byte, gameId uint64, userAuthorizer blockchain.Authorizer) {
	commandType := "GAME_CREATE"

	uint8Merkle := byteArrayToUint(rootMerkel)

	payload := []any{
		stake,
		uint8Merkle,
		gameId,
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
	userAuthorizer blockchain.Authorizer,
) {
	commandType := "GAME_MOVE"
	payload := []any{
		gameId,
		guessX,
		guessY,
		proof,
		blockPresent,
		opponentGuessX,
		opponentGuessY,
		nonce,
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
		PlayedAt:    time.Now().UTC().UnixMilli(),
	}

	result := b.db.Create(&mh)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling Moved")
		return
	}

	message.Ack()

	var isHit bool
	result = b.db.
		Raw(`
			SELECT EXISTS(
				SELECT 1 
				FROM game_grid_point 
				WHERE game_id = ? 
				  AND coordinate_x = ? 
				  AND coordinate_y = ? 
				  AND block_present = true);
        `, messagePayload.GameId, messagePayload.X, messagePayload.Y).
		Scan(&isHit)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Cannot fetch isHit for player move")
		// should have proper ws error signal implemented
		// but not necessary for this poc
		isHit = false
	}

	wsEvent := map[string]any{
		"type": "MOVE_DONE",
		"payload": map[string]any{
			"gameId": messagePayload.GameId,
			"userId": messagePayload.PlayerId,
			"x":      messagePayload.X,
			"y":      messagePayload.Y,
			"isHit":  isHit,
		},
	}
	b.notificationHub.Publish(fmt.Sprintf("game/%d", messagePayload.GameId), wsEvent)
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
		FlowId:      &messagePayload.Payload,
		OwnerId:     messagePayload.CreatorId,
		GameStatus:  "CREATED",
		Stake:       messagePayload.Stake,
		TimeCreated: timeNow.UnixMilli(),
	}

	result := b.db.Create(&game)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling GameOver")
		return
	}

	message.Ack()

	wsEvent := map[string]any{
		"type": "GAME_CREATED",
		"payload": map[string]any{
			"gameId":    messagePayload.GameId,
			"creatorId": messagePayload.CreatorId,
			"stake":     messagePayload.Stake,
		},
	}
	b.notificationHub.Publish(fmt.Sprintf("game/%d", messagePayload.GameId), wsEvent)
}

func (b *gameContractBridge) handleChallengerJoined(_ context.Context, _ *gcppubsub.Message) {
	// TODO when joined command done
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

	wsEvent := map[string]any{
		"type": "GAME_OVER",
		"payload": map[string]any{
			"gameId":   messagePayload.GameId,
			"winnerId": messagePayload.Winner,
		},
	}
	b.notificationHub.Publish(fmt.Sprintf("game/%d", messagePayload.GameId), wsEvent)
}

func byteArrayToUint(data []byte) []uint64 {
	// create buffer from byte array
	buffer := bytes.NewReader(data)

	// read bytes from buffer as uint64
	uint8Data := make([]uint64, len(data))
	for i := 0; i < len(uint8Data); i++ {
		var num uint8
		binary.Read(buffer, binary.BigEndian, &num)
		uint8Data[i] = uint64(num)
	}

	return uint8Data
}
