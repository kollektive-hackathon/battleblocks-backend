package model

type User struct {
	Id                       uint64
	Email                    string
	Username                 string
	CustodialWalletId        uint64
	SelfCustodyWalletAddress string
	GoogleIdentityId         string
}
