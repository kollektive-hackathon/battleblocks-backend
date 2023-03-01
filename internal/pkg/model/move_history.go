package model

type MoveHistory struct {
	Id          uint64 `json:id`
	UserId      uint64 `json:userId`
	GameId      uint64 `json:gameId`
	Coordinatex uint   `json:coordinateX`
	Coordinatey uint   `json:coordinateY`
	PlayedAt    int64  `json:playedAt`
}
