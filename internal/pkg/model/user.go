package model

type User struct {
	Id                       uint64 `json:"id"`
	Email                    string `json:"email"`
	Username                 string `json:"username"`
	CustodialWalletId        uint64 `json:"custodialWalletId"`
	SelfCustodyWalletAddress string `json:"selfCustodyWalletAddress"`
	GoogleIdentityId         string `json:"googleIdentityId"`
}

func (User) TableName() string {
	return "battleblocks_user"
}
