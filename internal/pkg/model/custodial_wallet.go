package model

type CustodialWallet struct {
	Id         uint64 `gorm:"primaryKey" json:"id"`
	ResourceId string `json:"resourceId"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}

func (CustodialWallet) TableName() string {
	return "custodial_wallet"
}
