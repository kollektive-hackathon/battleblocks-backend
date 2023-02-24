package shop

import (
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
)

type nftTransactionService struct {
}

func (nts *nftTransactionService) mint(recipientAddress, block model.Block, authorizers []blockchain.Authorizer) {
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

func (nts *nftTransactionService) transfer(recipientAddress string, withdrawId uint64, authorizers []blockchain.Authorizer) {
	commandType := "NFT_TRANSFER"
	payload := []any{
		recipientAddress,
		withdrawId,
	}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (nts *nftTransactionService) transferAdmin(recipientAddress string, withdrawId uint64, authorizers []blockchain.Authorizer) {
	commandType := "NFT_TRANSFER_ADMIN"
	payload := []any{
		recipientAddress,
		withdrawId,
	}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}

func (nts *nftTransactionService) burn(id uint64, authorizers []blockchain.Authorizer) {
	commandType := "NFT_BURN"
	payload := []any{
		id,
	}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}
