package cosign

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type cosignService struct {
	db *gorm.DB
}

func (cs *cosignService) VerifyAndSign(credentials auth.Token, request CosignRequest) ([]byte, *reject.ProblemWithTrace) {
	signable, err := cs.parsePayload(request.Payload)
	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	log.
		Info().
		Msg(fmt.Sprintf("Checking cosign transaction %+v", signable))

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

	custodialWallet := cs.validate(transaction, credentials)

	if custodialWallet != nil {
		return cs.signVoucher(transaction, *custodialWallet)
	}

	err = fmt.Errorf("invalid request: you are not authorized to request this signature")
	return nil, &reject.ProblemWithTrace{
		Problem: reject.UnexpectedProblem(err),
		Cause:   err,
	}
}

func (cs *cosignService) validate(transaction *flow.Transaction, credentials auth.Token) *model.CustodialWallet {
	requestTxCode := string(transaction.Script)
	if cs.validateRequestTxCode(requestTxCode) {
		return nil
	}

	userGoogleId := credentials.Subject

	log.
		Debug().
		Msg(fmt.Sprintf("Fetching custodial wallet by user id %s and address %s", userGoogleId))

	var custodialWallet model.CustodialWallet
	result := cs.db.
		Model(&custodialWallet).
		Where("id = (SELECT custodial_wallet_id FROM battleblocks_user WHERE google_identity_id = ?)", userGoogleId).
		First(&custodialWallet)

	if result.Error != nil {
		return nil
	}

	return &custodialWallet
}

func (cs *cosignService) signVoucher(transaction *flow.Transaction, custodialWallet model.CustodialWallet) ([]byte, *reject.ProblemWithTrace) {
	accountKMSKey, err := cloudkms.KeyFromResourceID(custodialWallet.ResourceId)
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

	err = transaction.SignEnvelope(flow.HexToAddress(*custodialWallet.Address), 0, signer)

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

func (cs *cosignService) validateRequestTxCode(requestTxCode string) bool {
	pattern := regexp.MustCompile(`\s`)
	serverTxCode := cs.getTxCode()
	serverTxCode = pattern.ReplaceAllString(serverTxCode, "")
	requestTxCode = pattern.ReplaceAllString(requestTxCode, "")

	if serverTxCode == requestTxCode {
		return true
	}

	log.Warn().Msg(fmt.Sprintf(
		"Transactions dont match: server \"%s\", request \"%s\"",
		serverTxCode,
		requestTxCode))

	return false
}

func (cs *cosignService) getTxCode() string {
	txCode := cs.getTxCodeTemplate()

	addressTemplates := map[string]string{
		"BATTLE_BLOCKS_ACCOUNTS_ADDRESS": viper.Get("BATTLE_BLOCKS_ACCOUNTS_ADDRESS").(string),
		"BATTLE_BLOCKS_NFT_ADDRESS":      viper.Get("BATTLE_BLOCKS_NFT_ADDRESS").(string),
		"FUNGIBLE_TOKEN_ADDRESS":         viper.Get("FUNGIBLE_TOKEN_ADDRESS").(string),
		"NON_FUNGIBLE_TOKEN_ADDRESS":     viper.Get("NON_FUNGIBLE_TOKEN_ADDRESS").(string),
	}

	for k, v := range addressTemplates {
		txCode = strings.ReplaceAll(txCode, k, v)
	}

	return txCode
}

func (cs *cosignService) getTxCodeTemplate() string {
	return `import BattleBlocksAccounts from 0xBATTLE_BLOCKS_ACCOUNTS_ADDRESS
import BattleBlocksNFT from 0xBATTLE_BLOCKS_NFT_ADDRESS
import FungibleToken from 0xFUNGIBLE_TOKEN_ADDRESS
import NonFungibleToken from 0xNON_FUNGIBLE_TOKEN_ADDRESS


transaction {

    let authAccountCap: Capability<&AuthAccount>
    let managerRef: &BattleBlocksAccounts.BattleBlocksAccountManager
    let childRef: &BattleBlocksAccounts.BattleBlocksAccount

    prepare(parent: AuthAccount, child: AuthAccount) {
        
        /* --- Configure parent's BattleBlocksAccountManager --- */
        //
        // Get BattleBlocksAccountManager Capability, linking if necessary
        if parent.borrow<&BattleBlocksAccounts.BattleBlocksAccountManager>(from: BattleBlocksAccounts.BattleBlocksAccountManagerStoragePath) == nil {
            // Save
            parent.save(<-BattleBlocksAccounts.createBattleBlocksAccountManager(), to: BattleBlocksAccounts.BattleBlocksAccountManagerStoragePath)
        }
        // Ensure BattleBlocksAccountManagerPublic is linked properly
        if !parent.getCapability<&{BattleBlocksAccounts.BattleBlocksAccountManagerPublic}>(BattleBlocksAccounts.BattleBlocksAccountManagerPublicPath).check() {
            parent.unlink(BattleBlocksAccounts.BattleBlocksAccountManagerPublicPath)
            // Link
            parent.link<
                &{BattleBlocksAccounts.BattleBlocksAccountManagerPublic}
            >(
                BattleBlocksAccounts.BattleBlocksAccountManagerPublicPath,
                target: BattleBlocksAccounts.BattleBlocksAccountManagerStoragePath
            )
        }
        // Get a reference to the BattleBlocksAccountManager resource
        self.managerRef = parent
            .borrow<
                &BattleBlocksAccounts.BattleBlocksAccountManager
            >(
                from: BattleBlocksAccounts.BattleBlocksAccountManagerStoragePath
            )!

        /* --- Link the child account's AuthAccount Capability & assign --- */
        //
        // Get the AuthAccount Capability, linking if necessary
        if !child.getCapability<&AuthAccount>(BattleBlocksAccounts.AuthAccountCapabilityPath).check() {
            // Unlink any Capability that may be there
            child.unlink(BattleBlocksAccounts.AuthAccountCapabilityPath)
            // Link & assign the AuthAccount Capability
            self.authAccountCap = child.linkAccount(BattleBlocksAccounts.AuthAccountCapabilityPath)!
        } else {
            // Assign the AuthAccount Capability
            self.authAccountCap = child.getCapability<&AuthAccount>(BattleBlocksAccounts.AuthAccountCapabilityPath)
        }

        // Get a refernce to the child account
        self.childRef = child.borrow<
                &BattleBlocksAccounts.BattleBlocksAccount
            >(
                from: BattleBlocksAccounts.BattleBlocksAccountStoragePath
            ) ?? panic("Could not borrow reference to BattleBlocksAccountTag in account ".concat(child.address.toString()))


        /* --- Set up BattleBlocksNFT.Collection --- */
        //
        if parent.borrow<&BattleBlocksNFT.Collection>(from: BattleBlocksNFT.CollectionStoragePath) != nil {
            // Create & save it to the account
            parent.save(<-BattleBlocksNFT.createEmptyCollection(), to: BattleBlocksNFT.CollectionStoragePath)

            // Create a public capability for the collection
            parent.link<
                &BattleBlocksNFT.Collection{NonFungibleToken.Receiver, NonFungibleToken.CollectionPublic, BattleBlocksNFT.BattleBlocksNFTCollectionPublic}
                >(
                    BattleBlocksNFT.CollectionPublicPath,
                    target: BattleBlocksNFT.CollectionStoragePath
                )

            // Link the Provider Capability in private storage
            parent.link<
                &BattleBlocksNFT.Collection{NonFungibleToken.Provider}
                >(
                    BattleBlocksNFT.ProviderPrivatePath,
                    target: BattleBlocksNFT.CollectionStoragePath
                )
        }
    }

    execute {
        // Add child account if it's parent-child accounts aren't already linked
        let childAddress = self.authAccountCap.borrow()!.address
        if !self.managerRef.getBattleBlocksAccountAddresses().contains(childAddress) {
            // Add the child account
            self.managerRef.addAsBattleBlocksAccount(battleBlocksAccountCap: self.authAccountCap, battleBlocksAccount: self.childRef)
        }
    }
}`
}
