package game

import "github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"

type JoinGameRequest struct {
	Placements []model.Placement
}
