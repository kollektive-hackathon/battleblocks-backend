package model

type GameStatus string

const (
	GameCreated   GameStatus = "CREATED"
	GamePreparing GameStatus = "PREPARING"
	GamePlaying   GameStatus = "PLAYING"
	GameFinished  GameStatus = "FINISHED"
)
