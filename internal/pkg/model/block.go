package model

type Block struct {
	Id        uint64 `json:"id"`
	Name      string `json:"name"`
	BlockType string `json:"blockType"`
	Rarity    string `json:"rarity"`
	Price     uint32 `json:"price"`
	ColorHex  string `json:"colorHex"`
}

func (Block) TableName() string {
	return "block"
}
