package profile

type Profile struct {
	Id                       uint64
	Email                    string
	Username                 string
	CustodialWalletAddress   string
	SelfCustodyWalletAddress string
	InventoryBlocks          []UserInventoryBlock `gorm:"embedded,embeddedPrefix:block_"`
}

type UserInventoryBlock struct {
	Id     uint64
	Name   string
	Type   string
	Rarity string
	Active bool
}
