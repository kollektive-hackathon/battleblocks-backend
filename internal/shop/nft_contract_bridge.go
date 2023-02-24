package shop

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"context"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
)

type nftContractBridge struct {
}

func (b *nftContractBridge) mint(recipientAddress, block model.Block, authorizers []blockchain.Authorizer) {
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

func (b *nftContractBridge) handleMinted(_ context.Context, _ *gcppubsub.Message) {

}

func (b *nftContractBridge) handleDeposited(_ context.Context, _ *gcppubsub.Message) {

}

func (b *nftContractBridge) handleBurned(_ context.Context, _ *gcppubsub.Message) {

}
