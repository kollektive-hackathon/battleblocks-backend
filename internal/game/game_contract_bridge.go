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
	PlayerAddress string `json:"playerAddress"`
	Turn          uint64 `json:"turn"`
	X             uint   `json:"coordinateX"`
	Y             uint   `json:"coordinateY"`
}

type ChallengerJoined struct {
	GameId    uint64 `json:"gameId"`
	StartTime uint64 `json:"startTime"`
	// Wager             float64          `json:"wager"`
	PlayerA *string `json:"playerA"`
	PlayerB *string `json:"playerB"`
	// Winner            *string           `json:"winner"`
	// PlayerHitCount    map[string]uint8 `json:"playerHitCount"`
	// GameState         uint8            `json:"gameState"`
	Turn uint8 `json:"turn"`
	// PlayerAMerkleRoot []uint8          `json:"playerAMerkleRoot"`
	// PlayerBMerkleRoot []uint8          `json:"playerBMerkleRoot"`
}

type GameOver struct {
	GameId         uint64          `json:"gameID"`
	PlayerA        string          `json:"playerA"`
	PlayerB        string          `json:"playerB"`
	Winner         string          `json:"winner"`
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
		gameId,
		stake,
		uint8Merkle,
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
	proof [][]uint8,
	blockPresent *bool,
	opponentGuessX *uint64,
	opponentGuessY *uint64,
	nonce *uint64,
	userAuthorizer blockchain.Authorizer,
) {
	var uint64Proof [][]uint64
	if proof != nil {
		uint64Proof = twoDimensionalbyteArrayToTwoDimensionalUint64Array(proof)
	}

	commandType := "GAME_MOVE"
	payload := []any{
		gameId,
		guessX,
		guessY,
		uint64Proof,
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

	er := b.db.Transaction(func(tx *gorm.DB) error {
		game, err := b.findGameByFlowID(messagePayload.GameId)
		result := tx.
			Model(&model.Game{}).
			Where("id = ?", messagePayload.GameId).
			Updates(map[string]any{
				"turn": messagePayload.Turn,
			})

		if result.Error != nil {
			log.Warn().Err(result.Error).Msg("Error while handling Moved")
			return result.Error
		}

		if err != nil {
			log.Warn().Err(err).Msg("Error while sending ChallengerJoined ws message")
			return result.Error
		}

		var user model.User
		f := tx.Raw(`SELECT bu.* FROM battleblocks_user bu
			LEFT JOIN custodial_wallet cw ON bu.custodial_wallet_id = cw.id
			WHERE cw.address = ?`, messagePayload.PlayerAddress).First(&user)

		if f.Error != nil {
			log.Warn().Err(err).Msg("Error while handling moved message")
			return f.Error
		}

		mh := model.MoveHistory{
			UserId:      user.Id,
			GameId:      game.Id,
			Coordinatex: messagePayload.X,
			Coordinatey: messagePayload.Y,
			PlayedAt:    time.Now().UTC().UnixMilli(),
		}

		result = tx.Table("move_history").Create(&mh)

		if result.Error != nil {
			log.Warn().Err(result.Error).Msg("Error while handling Moved")
			return result.Error
		}

		var isHit bool
		result = tx.
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
				"userId": user.Id,
				"x":      messagePayload.X,
				"y":      messagePayload.Y,
			},
		}
		b.notificationHub.Publish(fmt.Sprintf("game/%d", game.Id), wsEvent)
		return nil
	})
	if er != nil {
		log.Warn().Err(err).Msg("Cannot handle move event")
		return
	}
	message.Ack()
}

func (b *gameContractBridge) handleGameCreated(_ context.Context, message *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(message.Data))
	messagePayload, err := utils.JsonDecodeByteStream[GameCreated](message.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Error while parsing GameCreated message")
		return
	}

	result := b.db.
		Model(&model.Game{}).
		Where("id = ?", messagePayload.Payload).
		Updates(map[string]any{
			"flow_id":     messagePayload.GameId,
			"game_status": "CREATED",
		})

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling GameCreated")
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
	b.notificationHub.Publish(fmt.Sprintf("game/%d", messagePayload.Payload), wsEvent)
}

func (b *gameContractBridge) handleChallengerJoined(_ context.Context, m *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(m.Data))
	messagePayload, err := utils.JsonDecodeByteStream[ChallengerJoined](m.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Error while parsing ChallengedJoined message")
		return
	}
	if messagePayload.PlayerB != nil {
		b.db.Transaction(func(tx *gorm.DB) error {
			var user model.User
			f := tx.Raw(`SELECT bu.* FROM battleblocks_user bu
				LEFT JOIN custodial_wallet cw ON bu.custodial_wallet_id = cw.id
				WHERE cw.address = ?`, messagePayload.PlayerB).First(&user)

			if f.Error != nil {
				log.Warn().Err(err).Msg("Error while handling ChallengerJoined message")
				return f.Error
			}

			f = tx.
				Model(&model.Game{}).
				Where("flow_id = ?", messagePayload.GameId).
				Updates(map[string]any{
					"challenger_id": user.Id,
					"game_status":   "PLAYING",
					"turn":          messagePayload.Turn,
				})

			if f.Error != nil {
				log.Warn().Err(err).Msg("Error while handling ChallengerJoined message")
				return f.Error
			}

			game, err := b.findGameByFlowID(messagePayload.GameId)
			if err != nil {
				log.Warn().Err(err).Msg("Error while sending ChallengerJoined ws message")
				return err
			}
			wsEvent := map[string]any{
				"type": "CHALLENGER_JOINED",
				"payload": map[string]any{
					"challengerName": user.Username,
					"turn":           messagePayload.Turn,
					"gameStatus":     "PLAYING",
				},
			}

			b.notificationHub.Publish(fmt.Sprintf("game/%d", game.Id), wsEvent)

			return nil
		})

	}
	m.Ack()

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
		Where("flow_id = ?", messagePayload.GameId).
		Updates(map[string]any{
			"winner":      messagePayload.Winner,
			"game_status": "FINISHED",
		})

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling GameOver")
		return
	}

	game, err := b.findGameByFlowID(messagePayload.GameId)
	if err != nil {
		log.Warn().Err(err).Msg("Error while sending ChallengerJoined ws message")
		return
	}

	message.Ack()

	if err != nil {
		log.Warn().Err(result.Error).Msg("Error while sending ChallengerJoined ws message")
		return
	}

	wsEvent := map[string]any{
		"type": "GAME_OVER",
		"payload": map[string]any{
			"gameId":   game.Id,
			"winnerId": messagePayload.Winner,
		},
	}
	b.notificationHub.Publish(fmt.Sprintf("game/%d", game.Id), wsEvent)
}

func (b *gameContractBridge) findGameByFlowID(flowID uint64) (model.Game, error) {
	var game model.Game
	result := b.db.Model(&model.Game{}).
		Where("flow_id = ?", flowID).
		First(&game)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while fetching game by flow ID")
		return game, result.Error
	}

	return game, nil
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

func twoDimensionalbyteArrayToTwoDimensionalUint64Array(data [][]byte) [][]uint64 {
	uint64Data := make([][]uint64, len(data))
	for i, row := range data {
		uint64Data[i] = make([]uint64, len(row))
		buffer := bytes.NewReader(row)
		for j := 0; j < len(row); j++ {
			var num uint8
			binary.Read(buffer, binary.BigEndian, &num)
			uint64Data[i][j] = uint64(num)
		}
	}
	return uint64Data
}
