package cosign

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"strconv"
)

type cosignService struct {
	db *gorm.DB
}

func (cs *cosignService) VerifyAndSign(credentials auth.Token, request CosignRequest) ([]byte, *reject.ProblemWithTrace) {
	transaction, err := cs.parsePayload(request.Payload)
	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	log.
		Info().
		Msg(fmt.Sprintf("Checking cosign transaction %+v", transaction))

	if cs.validate(transaction, credentials) == true {
		return cs.signVoucher(transaction)
	}

	err = fmt.Errorf("invalid request: you are not authorized to request this signature")
	return nil, &reject.ProblemWithTrace{
		Problem: reject.UnexpectedProblem(err),
		Cause:   err,
	}
}

func (cs *cosignService) validate(transaction *Signable, credentials auth.Token) bool {

	serverTxCode := "" // TODO add code string here
	requestTxCode := transaction.Cadence

	if serverTxCode != requestTxCode {
		return false
	}

	userGoogleId := credentials.Subject
	address := transaction.Args[0] // TODO check which arg is address

	var custodialWallet model.CustodialWallet
	result := cs.db.
		Model(&custodialWallet).
		Where("address = ? AND id = (SELECT custodial_wallet_id FROM user WHERE google_identity_id = ?)", address, userGoogleId).
		First(&custodialWallet)

	if result.Error != nil {
		return false
	}

	return true
}

func (cs *cosignService) signVoucher(signable *Signable) ([]byte, *reject.ProblemWithTrace) {
	decodedData, err := hex.DecodeString(signable.Message[64:])

	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	transaction, err := flow.DecodeTransaction(decodedData)

	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	// TODO umjesto signa adminom uvijek - signamo s userovima resourceIdEM-custody-w
	tfcAdminAddress, tfcAdminKeyIndex, gcpKmsResourceName := cs.getCosignEnv()

	accountKMSKey, err := cloudkms.KeyFromResourceID(gcpKmsResourceName)
	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	ctx := context.Background()
	kmsClient, err := cloudkms.NewClient(ctx)
	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	signer, err := kmsClient.SignerForKey(
		ctx,
		accountKMSKey,
	)
	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	err = transaction.SignEnvelope(flow.HexToAddress(tfcAdminAddress), tfcAdminKeyIndex, signer)

	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	return transaction.EnvelopeSignatures[len(transaction.EnvelopeSignatures)-1].Signature, nil
}

func (cs *cosignService) parsePayload(payload map[string]any) (*Signable, error) {
	jsonByteSlice, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var t Signable
	err = json.Unmarshal(jsonByteSlice, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (cs *cosignService) getCosignEnv() (string, int, string) {
	tfcAdminAddress := viper.Get("TFC_ADMIN_ADDRESS").(string)

	var err error
	tfcAdminKeyIndex, err := strconv.Atoi(viper.Get("TFC_ADMIN_KEY_INDEX").(string))
	gcpKmsResourceName := viper.Get("GCP_KMS_RESOURCE_NAME").(string)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing TFC_ADMIN_KEY_INDEX")
	}

	return tfcAdminAddress, tfcAdminKeyIndex, gcpKmsResourceName
}
