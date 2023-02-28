package model

type Game struct {
	Id           uint64     `json:"id"`
	FlowId       *uint64    `json:"flowId"`
	OwnerId      uint64     `json:"ownerId"`
	ChallengerId *uint64    `json:"challengerId"`
	GameStatus   GameStatus `json:"gameStatus"`
	Stake        uint64     `json:"stake"`
	TimeStarted  int64      `json:"timeStarted"`
	TimeCreated  int64      `json:"timeCreated"`
	WinnerId     *uint64    `json:"winnerId"`
	Turn         *uint64    `json:"turn"`
}

func (Game) TableName() string {
	return "game"
}
