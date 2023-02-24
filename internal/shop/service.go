package shop

import (
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
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
