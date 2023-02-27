package shop

import (
	"errors"

	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type shopService struct {
	db     *gorm.DB
	bridge *nftContractBridge
}

func (ss *shopService) FindAll() ([]model.Block, *reject.ProblemWithTrace) {
	var blocks []model.Block
	result := ss.db.
		Model(&model.Block{}).
		Where("stock = false").
		Find(&blocks)

	if result.Error != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	return blocks, nil
}

func (ss *shopService) SendBoughtToUser(userId, blockId string) {
	ss.db.Transaction(func(tx *gorm.DB) error {
		var block model.Block
		result := ss.db.
			Model(&model.Block{}).
			Where("id = ?", blockId).
			First(&block)

		if result.Error != nil {
			log.Warn().Interface("er", result.Error.Error()).Msg("could not get block to mint")
			return result.Error
		}

		var wallet model.CustodialWallet
		result = tx.Raw(`SELECT cw FROM battleblocks_user bu
			LEFT JOIN custodial_wallet cw ON bu.custodial_wallet_id = cw.id 
			WHERE bu.id = ?`, userId).
			First(wallet)

		if result.Error != nil {
			log.Warn().Interface("er", result.Error.Error()).Msg("error fetching address of user")
			return result.Error
		}

		ss.bridge.mint(*wallet.Address, block, []blockchain.Authorizer{blockchain.GetAdminAuthorizer()})
		return nil
	})
}
