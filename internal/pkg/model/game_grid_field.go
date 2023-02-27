package model

type GameGridField struct {
	GameId       uint64
	UserId       uint64
	BlockPresent bool
	CoordinateX  uint64
	CoordinateY  uint64
	Nonce        string
}