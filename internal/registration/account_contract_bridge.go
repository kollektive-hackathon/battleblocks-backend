package registration

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/ws"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/profile"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type AccountCreated struct {
	PublicKey string `json:"originatingPublicKey"`
	Address   string `json:"address"`
}

type AccountDelegated struct {
	PublicKey           string `json:"originatingPublicKey"`
	CustodialAddress    string `json:"address"`
	NonCustodialAddress string `json:"parent"`
}

type accountContractBridge struct {
	db              *gorm.DB
	notificationHub *ws.WebSocketNotificationHub
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

	var p profile.Profile
	result = b.db.
		Table("battleblocks_user").
		Joins("INNER JOIN user_block_inventory ON battleblocks_user.id = user_block_inventory.user_id").
		Joins("INNER JOIN custodial_wallet ON battleblocks_user.custodial_wallet_id = custodial_wallet.id").
		Where("custodial_wallet.address = ?", messagePayload.Address).
		Select(`
			battleblocks_user.id, 
			battleblocks_user.email,
			battleblocks_user.username,
			custodial_wallet.address AS custodial_wallet_address,
			battleblocks_user.self_custody_wallet_address AS self_custody_wallet_address
		`).
		Scan(&p)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Cannot fetch profile on AccountCreated event")
	}

	b.notificationHub.Publish(fmt.Sprintf("registration/%s", p.Email), p)
}

func (b *accountContractBridge) handleCustodialAccountDelegated(_ context.Context, message *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(message.Data))
	messagePayload, err := utils.JsonDecodeByteStream[AccountDelegated](message.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Error while parsing AccountDelegated message")
		return
	}

	result := b.db.
		Model(&model.User{}).
		Where("custodial_wallet_id = (SELECT id FROM custodial_wallet WHERE address = ?)", messagePayload.CustodialAddress).
		Update("self_custody_wallet_address", messagePayload.NonCustodialAddress)

	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("Error while handling AccountCreated")
		return
	}

	message.Ack()
}
