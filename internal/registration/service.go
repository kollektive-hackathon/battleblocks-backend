package registration

import (
	"context"
	"fmt"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/keymgmt"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"gorm.io/gorm"
)

type registrationService struct {
	db     *gorm.DB
	bridge *accountContractBridge
}

func (s *registrationService) register(username string, email string, googleIdentityId string) *reject.ProblemWithTrace {
	ctx := context.Background()
	defaultKeyIndex := 0
	defaultKeyWeight := -1
	accountKey, _, rid, err := keymgmt.GenerateAsymetricKey(ctx, defaultKeyIndex, defaultKeyWeight)
	if err != nil {
		return &reject.ProblemWithTrace{Problem: reject.UnexpectedProblem(err), Cause: err}
	}

	publicKey := accountKey.PublicKey.String()

	err = s.db.Transaction(func(tx *gorm.DB) error {
		cw := model.CustodialWallet{
			ResourceId: *rid,
			PublicKey:  publicKey,
		}
		result := tx.Create(&cw)
		if result.Error != nil {
			return result.Error
		}

		user := model.User{
			Email:                    email,
			Username:                 username,
			CustodialWalletId:        cw.Id,
			SelfCustodyWalletAddress: "",
			GoogleIdentityId:         googleIdentityId,
		}
		result = tx.Create(&user)
		if result.Error != nil {
			return result.Error
		}

		result = tx.Exec(fmt.Sprintf(`INSERT INTO user_block_inventory(user_id, block_id, active)
                            SELECT %d, id, true FROM block WHERE stock = true`, user.Id))

		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return &reject.ProblemWithTrace{Problem: reject.UnexpectedProblem(err), Cause: err}
	}

	s.bridge.createCustodialAccount(publicKey)

	return nil
}
