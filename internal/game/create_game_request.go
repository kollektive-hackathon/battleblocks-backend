package game

import "github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"

type CreateGameRequest struct {
	Stake      float32           `json:"stake"`
	Placements []model.Placement `json:"placements"`
}
