package shop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gcppubsub "cloud.google.com/go/pubsub"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type MintedEvent struct {
	To   string `json:"to"`
	Name string `json:"name"`
	Id   uint64 `json:"id"`
}

type nftContractBridge struct {
	db *gorm.DB
}

func (b *nftContractBridge) mint(recipientAddress string, block model.Block, authorizers []blockchain.Authorizer) {
	commandType := "NFT_MINT"
	payload := []any{
		recipientAddress,
		block.Name,
		map[string]any{
			"type":     block.BlockType,
			"rarity":   block.Rarity,
			"colorHex": block.ColorHex,
		},
	}

	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (b *nftContractBridge) transfer(recipientAddress string, withdrawId uint64, authorizers []blockchain.Authorizer) {
	commandType := "NFT_TRANSFER"
	payload := []any{
		recipientAddress,
		withdrawId,
	}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (b *nftContractBridge) transferAdmin(recipientAddress string, withdrawId uint64, authorizers []blockchain.Authorizer) {
	commandType := "NFT_TRANSFER_ADMIN"
	payload := []any{
		recipientAddress,
		withdrawId,
	}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (b *nftContractBridge) burn(id uint64, authorizers []blockchain.Authorizer) {
	commandType := "NFT_BURN"
	payload := []any{
		id,
	}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

// TODO implement consumers
func (b *nftContractBridge) handleWithdrew(_ context.Context, _ *gcppubsub.Message) {
}

func (b *nftContractBridge) handleMinted(_ context.Context, m *gcppubsub.Message) {
	log.Info().Msg("Received message payload " + string(m.Data))

	var eventData MintedEvent
	err := json.Unmarshal(m.Data, &eventData)

	if err != nil {
		log.Warn().Err(err).Msg("Could not unmarshal minted event data")
		return
	}

	err = b.db.Transaction(func(tx *gorm.DB) error {
		var block model.Block
		f := tx.Table("block").Where("name = ?", eventData.Name).First(&block)
		if f.Error != nil {
			log.Warn().Msg("error fetching block to transfer to user")
			return errors.New("error fetching block to transfer to user")
		}

		var user model.User
		f = tx.Raw(`SELECT bu.* FROM
			battleblocks_user bu LEFT JOIN
			custodial_wallet cw ON bu.custodial_wallet_id = cw.id
			WHERE cw.address = ?`, eventData.To).First(&user)
		if f.Error != nil {
			log.Warn().Msg("error fetching user to transfer to ")
			return errors.New("error fetching user to transfer to ")
		}

		// Insert NFT on user with TO
		// Insert NFT purchase history
		flowId := eventData.Id
		nft := model.Nft{
			FlowId:  flowId,
			BlockId: block.Id,
		}
		tx.Table("nft").Create(&nft)

		nft_history := model.NftPurchaseHistory{
			NftId:       nft.Id,
			BuyerId:     user.Id,
			PurchasedAt: time.Now().UTC().UnixMilli(),
		}

		tx.Table("nft_purchase_history").Create(&nft_history)

		result := tx.Exec(fmt.Sprintf(`INSERT INTO user_block_inventory(user_id, block_id, active)
			SELECT %d, %d, true`, user.Id, block.Id))

		if result.Error != nil {
			log.Warn().Msg("error inserting the item to user inventory")
			return errors.New("error inserting the item to user inventory")
		}

		return nil
	})
	if err != nil {
		log.Info().Interface("ev", eventData).Msg("Could not handle mitned event")
		return
	}
	m.Ack()
}

func (b *nftContractBridge) handleDeposited(_ context.Context, _ *gcppubsub.Message) {

}

func (b *nftContractBridge) handleBurned(_ context.Context, _ *gcppubsub.Message) {

}
