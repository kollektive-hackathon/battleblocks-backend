package game

import (
	"net/http"

	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	databaseError = "error.data.access"
)

type gameService struct {
	db *gorm.DB
}

func (gs *gameService) getGames(page utils.PageRequest, userId string) ([]model.Game, *int64, *reject.ProblemWithTrace) {
	games := []model.Game{}
	gamesSize := int64(0)

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table("game").
			Count(&gamesSize)
		if res.Error != nil {
			return res.Error
		}

		res = tx.Table("game").
			Limit(page.Size).
			Offset(page.Offset).
			Clauses(clause.OrderBy{
				Expression: clause.Expr{
					SQL:
					"(owner_id = $1 AND game_status = 'PLAYING') DESC, (owner_id = $1) DESC, time_created DESC",
					Vars: []interface{}{userId},
				WithoutParentheses: true,
				},
			}).
			Scan(&games)
		if res.Error != nil {
			return res.Error
		}
		return nil
	})

	if err != nil {
		return nil, nil, &reject.ProblemWithTrace{
			Problem: reject.NewProblem().
				WithTitle("Trouble fetching data from database").
				WithStatus(http.StatusInternalServerError).
				WithCode(databaseError).
				Build(),
			Cause: err,
		}

	}
	return games, &gamesSize, nil
}

func (gs *gameService) createGame(game model.Game) *reject.ProblemWithTrace {
	// TODO Send tx for creating game with the model
	return nil
}
