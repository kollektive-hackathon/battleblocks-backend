package registration

import (
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/pubsub"
)

type accountTransactionsService struct {
}

func (ats *accountTransactionsService) createCustodialAccount(publicKey string) {
	commandType := "CREATE_USER_ACCOUNT"
	initialFundingAmount := float64(10)
	payload := []any{
		publicKey,
		initialFundingAmount,
	}
	authorizers := []blockchain.Authorizer{blockchain.GetAdminAuthorizer()}
	cmd := blockchain.NewBlockchainCommand(commandType, payload, authorizers)
	pubsub.Publish(cmd)
}
