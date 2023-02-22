package registration

import (
	"context"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/keymgmt"
	"github.com/kollektive-hackathon/battleblocks-backend/pkg/reject"
	"gorm.io/gorm"
)

type registrationService struct {
	db *gorm.DB
}

func (s registrationService) register(username string, email string, googleIdentityId string) *reject.ProblemWithTrace {

	// TODO create kms keymgmt
	ctx := context.Background()
	defaultKeyIndex := 0
	defaultKeyWeight := -1
	accountKey, privateKey, err := keymgmt.GenerateAsymetricKey(ctx, defaultKeyIndex, defaultKeyWeight)
	if err != nil {
		return nil
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {

		// TODO create custodialWallet

		result := s.db.Exec(
			`INSERT INTO user(email, username, custodial_wallet_id, self_custody_wallet_address, google_identity_id)
                   VALUES (?, ?, ?, null, ?)`, email, username, cc, googleIdentityId)

		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return &reject.ProblemWithTrace{Problem: reject.UnexpectedProblem(err), Cause: err}
	}

	return nil
}
