package model

type CustodialWallet struct {
	ID         uint64  `gorm:"primaryKey" json:"id"`
	ResourceId string  `json:"resourceId"`
	PublicKey  string  `json:"publicKey"`
	Address    *string `json:"address"`
}

func (CustodialWallet) TableName() string {
	return "custodial_wallet"
}
