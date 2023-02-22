package registration

import (
	"context"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/keymgmt"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	reject2 "github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"gorm.io/gorm"
)

type registrationService struct {
	db *gorm.DB
}

func (s registrationService) register(username string, email string, googleIdentityId string) *reject2.ProblemWithTrace {
	ctx := context.Background()
	defaultKeyIndex := 0
	defaultKeyWeight := -1
	accountKey, _, err := keymgmt.GenerateAsymetricKey(ctx, defaultKeyIndex, defaultKeyWeight)
	if err != nil {
		return nil
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// TODO test with gcp and pass/save resourceID here from GenerateAsymetricKey

		cw := model.CustodialWallet{
			ResourceId: "resourceID-TODO",
			PublicKey:  accountKey.PublicKey.String(),
			Address:    "",
		}
		result := s.db.Create(cw)
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
		result = s.db.Create(user)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return &reject2.ProblemWithTrace{Problem: reject2.UnexpectedProblem(err), Cause: err}
	}

	// TODO pubsub tx service for creating custodial wallet address

	return nil
}
