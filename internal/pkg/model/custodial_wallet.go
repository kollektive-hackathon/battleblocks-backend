package model

type CustodialWallet struct {
	Id         uint64 `gorm:"primaryKey"`
	ResourceId string
	PublicKey  string
	Address    string
}
