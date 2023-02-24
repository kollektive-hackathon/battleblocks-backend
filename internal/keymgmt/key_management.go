package keymgmt

import (
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/jpillora/backoff"
	"strings"
	"time"
)

type PrivateKey struct {
	Index    int                       `json:"index"`
	Type     string                    `json:"type"`
	Value    string                    `json:"-"`
	SignAlgo crypto.SignatureAlgorithm `json:"-"`
	HashAlgo crypto.HashAlgorithm      `json:"-"`
}

func GenerateAsymetricKey(ctx context.Context, keyIndex, weight int) (*flow.AccountKey, *PrivateKey, error) {
	u := uuid.New()

	googleKmsProjectId := viper.Get("GOOGLE_KMS_PROJECT_ID").(string)
	googleKmsLocationId := viper.Get("GOOGLE_KMS_LOCATION_ID").(string)
	googleKmsKeyRingId := viper.Get("GOOGLE_KMS_KEYRING_ID").(string)

	// Create the new key in Google KMS
	k, err := createAsymetricKey(
		ctx,
		fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", googleKmsProjectId, googleKmsLocationId, googleKmsKeyRingId),
		fmt.Sprintf("battleblocks-custodial-wallet-account-key-%s", u.String()),
	)
	if err != nil {
		return nil, nil, err
	}

	client, err := cloudkms.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	pub, h, s, err := GetPublicKey(ctx, client, k)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("failed to get public key for Google KMS key, keyId: %s", k.KeyID))
		return nil, nil, err
	}

	f := flow.NewAccountKey().
		SetPublicKey(*pub).
		SetHashAlgo(*h).
		SetWeight(weight)
	f.Index = keyIndex

	p := &PrivateKey{
		Index:    keyIndex,
		Type:     "google_kms",
		Value:    k.ResourceID(),
		SignAlgo: *s,
		HashAlgo: *h,
	}

	return f, p, nil
}

func GetPublicKey(ctx context.Context, kmsClient *cloudkms.Client, kmsKey *cloudkms.Key) (*crypto.PublicKey, *crypto.HashAlgorithm, *crypto.SignatureAlgorithm, error) {
	// Get the public key (using flow-go-sdk's cloudkms.Client)
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    time.Minute,
		Factor: 5,
		Jitter: true,
	}

	deadline := time.Now().Add(60 * time.Second)

	var publicKey crypto.PublicKey
	var hashAlgo crypto.HashAlgorithm
	var signAlgo crypto.SignatureAlgorithm
	var err error

	log.Trace().Msg(fmt.Sprintf("Getting public key for KMS key, keyId: %s", kmsKey.KeyID))

	for {
		publicKey, hashAlgo, err = kmsClient.GetPublicKey(ctx, *kmsKey)
		if publicKey != nil {
			signAlgo = publicKey.Algorithm()
			break
		}
		// non-retryable error
		if err != nil && !strings.Contains(err.Error(), "KEY_PENDING_GENERATION") {
			//entry.WithFields(log.Fields{"err": err}).Error("failed to get public key")
			return nil, nil, nil, err
		}
		// key not generated yet, retry
		if err != nil && strings.Contains(err.Error(), "KEY_PENDING_GENERATION") {
			log.Trace().Msg("KMS key is pending creation, will retry")
			continue
		}

		time.Sleep(b.Duration())

		if time.Now().After(deadline) {
			err = fmt.Errorf("timeout while trying to get public key")
			log.Error().Err(err)
			return nil, nil, nil, err
		}
	}

	return &publicKey, &hashAlgo, &signAlgo, err
}

// AsymKey creates a new asymmetric signing key in Google KMS and returns
// a cloudkms.Key
func createAsymetricKey(ctx context.Context, parent string, id string) (*cloudkms.Key, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	r := &kmspb.CreateCryptoKeyRequest{
		Parent:      parent,
		CryptoKeyId: id,
		CryptoKey: &kmspb.CryptoKey{
			Purpose: kmspb.CryptoKey_ASYMMETRIC_SIGN,
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				Algorithm: kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256,
			},
			// TODO any labels?
			/*Labels: map[string]string{
				"service":         "battleblocks-api",
				"account_address": "",
				"chain_id":        "",
				"environment":     "",
			},*/
		},
	}

	gk, err := client.CreateCryptoKey(ctx, r)
	if err != nil {
		return nil, err
	}

	// Append cryptoKeyVersions so that we can utilize the KeyFromResourceID method
	k, err := cloudkms.KeyFromResourceID(fmt.Sprintf("%s/cryptoKeyVersions/1", gk.Name))
	if err != nil {
		return nil, err
	}

	// Validate key name
	if !strings.HasPrefix(k.ResourceID(), gk.Name) {
		err := fmt.Errorf("WARNING: created Google KMS key name does not match the expected")
		return nil, err
	}

	return &k, nil
}
