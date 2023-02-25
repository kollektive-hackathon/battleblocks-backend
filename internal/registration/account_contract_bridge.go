package registration

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"context"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type AccountCreated struct {
	PublicKey string `json:"originatingPublicKey"`
	Address   string `json:"address"`
}

type accountContractBridge struct {
	db *gorm.DB
}

func (b *accountContractBridge) createCustodialAccount(publicKey string) {
	commandType := "CREATE_USER_ACCOUNT"
	initialFundingAmount := float64(10)
	payload := []any{
		publicKey,
		initialFundingAmount,
	}
	authorizers := []blockchain.Authorizer{blockchain.GetAdminAuthorizer()}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (b *accountContractBridge) handleCustodialAccountCreated(_ context.Context, message *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(message.Data))
	messagePayload, err := utils.JsonDecodeByteStream[AccountCreated](message.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Error while parsing AccountCreated message")
		return
	}

	result := b.db.
		Model(&model.CustodialWallet{}).
		Where("public_key = ?", messagePayload.PublicKey).
		Update("address", messagePayload.Address)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling AccountCreated")
		return
	}

	message.Ack()
}
