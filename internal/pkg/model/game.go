package model

import (
	"time"
)

type Game struct {
	Id               uint64
	OwnerId          uint64
	ChallengerId     *uint64
	GameStatus       GameStatus
	Stake            uint64
	TimeStarted      *time.Time
	TimeCreated      *time.Time
	TimeLimitSeconds uint64
	WinnerId         *uint64
}
