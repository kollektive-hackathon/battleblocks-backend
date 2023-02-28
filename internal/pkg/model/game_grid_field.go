package model

type GameGridPoint struct {
	GameId       uint64 `gorm:"primaryKey"`
	UserId       uint64 `gorm:"primaryKey"`
	BlockPresent bool
	CoordinateX  uint64 `gorm:"primaryKey"`
	CoordinateY  uint64 `gorm:"primaryKey"`
	Nonce        string
}
