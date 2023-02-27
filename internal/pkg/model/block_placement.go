package model

type BlockPlacement struct {
	Id          uint64
	UserId      uint64
	GameId      uint64
	BlockId     string
	Coordinatex uint64
	Coordinatey uint64
}

func (BlockPlacement) TableName() string {
	return "block_placement"
}
