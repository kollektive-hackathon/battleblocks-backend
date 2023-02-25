package blockchain

import "github.com/spf13/viper"

func GetAdminAuthorizer() Authorizer {
	return Authorizer{
		KmsResourceId:        viper.Get("ADMIN_GCP_KMS_RESOURCE_NAME").(string),
		ResourceOwnerAddress: viper.Get("ADMIN_AUTHORIZER_ADDR").(string),
	}
}
