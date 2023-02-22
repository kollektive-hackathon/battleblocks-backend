package profile

import (
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"gorm.io/gorm"
)

type profileService struct {
	db *gorm.DB
}

func (s *profileService) FindById(id uint64) (*Profile, *reject.ProblemWithTrace) {
	var profile Profile
	result := s.db.
		Table("user").
		Joins("INNER JOIN user_block_inventory ON user.id = user_block_inventory.user_id").
		Joins("INNER JOIN block ON user_block_inventory.block_id = block.id").
		Joins("INNER JOIN custodial_wallet ON user.custodial_wallet_id = custodial_wallet.id").
		Where("user.id = ?", id).
		Select(`
			user.id, 
			user.email,
			user.username,
			custodial_wallet.address AS custodial_wallet_address,
			user.self_custody_wallet_address AS self_custody_wallet_address,
			block.id AS block_id,
			block.name AS block_name,
			block.block_type AS block_type,
			block.rarity AS block_rarity,
			user_block_inventory.active AS block_active
		`).
		Scan(&profile)

	if result.Error != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	return &profile, nil
}

func (s *profileService) activateBlocks(userId uint64, body ActivateBlocksRequest) *reject.ProblemWithTrace {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		result := s.db.Exec(
			`UPDATE user_block_inventory 
                    SET active = true 
                  WHERE block_id IN ? 
                    AND block_id = false
                    AND user_id = ?`, body.activeBlockIds, userId)
		if result.Error != nil {
			return result.Error
		}

		result = s.db.Exec(
			`UPDATE user_block_inventory 
                    SET active = false 
                  WHERE block_id NOT IN ? 
                    AND block_id = true
                    AND user_id = ?`, body.activeBlockIds, userId)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	return nil
}
