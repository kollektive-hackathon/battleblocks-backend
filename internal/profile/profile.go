package profile

type Profile struct {
	Id                       uint64               `json:"id"`
	Email                    string               `json:"email"`
	Username                 string               `json:"username"`
	CustodialWalletAddress   string               `json:"custodialWalletAddress"`
	SelfCustodyWalletAddress string               `json:"selfCustodyWalletAddress"`
	InventoryBlocks          []UserInventoryBlock `gorm:"-" json:"inventoryBlocks"`
}

type UserInventoryBlock struct {
	Id     uint64 `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Rarity string `json:"rarity"`
	Active bool   `json:"active"`
}
