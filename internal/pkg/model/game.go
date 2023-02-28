package model

type Game struct {
	Id           uint64
	FlowId       *uint64
	OwnerId      uint64
	ChallengerId *uint64
	GameStatus   GameStatus
	Stake        uint64
	TimeStarted  int64
	TimeCreated  int64
	WinnerId     *uint64
}

func (Game) TableName() string {
	return "game"
}
