package model

type MoveHistory struct {
	Id          uint64 `json:id`
	UserId      uint64 `json:userId`
	GameId      uint64 `json:gameId`
	CoordinateX uint   `json:coordinateX`
	CoordinateY uint   `json:coordinateY`
	PlayedAt    int64  `json:playedAt`
}
