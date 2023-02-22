package profile

type Profile struct {
	Id                       uint64               `json:"id"`
	Email                    string               `json:"email"`
	Username                 string               `json:"username"`
	CustodialWalletAddress   string               `json:"custodialWalletAddress"`
	SelfCustodyWalletAddress string               `json:"selfCustodyWalletAddress"`
	InventoryBlocks          []UserInventoryBlock `gorm:"embedded,embeddedPrefix:block_" json:"inventoryBlocks"`
}

type UserInventoryBlock struct {
	Id     uint64
	Name   string
	Type   string
	Rarity string
	Active bool
}
