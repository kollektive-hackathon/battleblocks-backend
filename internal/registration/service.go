package registration

import (
	"context"
	"fmt"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/keymgmt"
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
		// TODO refactor later so that auto generated ID is automatically set on struct - gorm.Create panics for some reason at the moment
		var walletId uint64
		result := tx.Exec(`INSERT INTO custodial_wallet(resource_id, public_key, address) VALUES (?, ?, null)`, *rid, publicKey)
		if result.Error != nil {
			return result.Error
		}

		result = tx.Raw("SELECT id FROM custodial_wallet WHERE resource_id = ?", *rid).Scan(&walletId)
		if result.Error != nil {
			return result.Error
		}

		result = tx.Exec(`INSERT INTO battleblocks_user(email, username, custodial_wallet_id, self_custody_wallet_address, google_identity_id) VALUES (?, ?, ?, null, ?)`,
			email, username, walletId, googleIdentityId)
		if result.Error != nil {
			return result.Error
		}

		var userId uint64
		result = tx.Raw("SELECT id FROM battleblocks_user WHERE username = ?", username).Scan(&userId)
		if result.Error != nil {
			return result.Error
		}

		result = tx.Exec(fmt.Sprintf(`INSERT INTO user_block_inventory(user_id, block_id, active)
                            SELECT %d, id, true FROM block WHERE stock = true`, userId))

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
