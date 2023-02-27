package game

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cbergoon/merkletree"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	databaseError = "error.data.access"
)

type TreeContent struct {
	field string
}

//CalculateHash hashes the values of a TestContent
func (t TreeContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.field)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

//Equals tests for equality of two Contents
func (t TreeContent) Equals(other merkletree.Content) (bool, error) {
	otherTC, ok := other.(TreeContent)
	if !ok {
		return false, errors.New("value is not of type TestContent")
	}
	return t.field == otherTC.field, nil
}

type gameService struct {
	db                 *gorm.DB
	gameContractBridge *gameContractBridge
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
					SQL:                "(owner_id = $1 AND game_status = 'PLAYING') DESC, (owner_id = $1) DESC, time_created DESC",
					Vars:               []interface{}{userId},
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
//TODO:
// Na svakom Moveu kreirati merkeltree ponovo iz block placementa i dobiti PROOF

func (gs *gameService) createGame(createGame CreateGameRequest, userId string) *reject.ProblemWithTrace {
	// TODO Send tx for creating game with the model
	// Spremiti u bazu game sa prizeom koji prima

	//placements also sent here
	//TODO: CHECK balance ---- execute script  with golang-flow-sdk
	// create merkeltree, send tx - with the ROOT and stake.

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var wallet model.CustodialWallet
		f := tx.Raw(`SELECT cw FROM battleblocks_user bu
			LEFT JOIN custodial_wallet cw ON bu.custodial_wallet_id = cw.id 
			WHERE bu.id = ?`, userId).
			First(wallet)

		if f.Error != nil {
			log.Warn().Msg("error fetching address of current user")
			return errors.New("error fetching address of current user")
		}

		balance, err := checkBalance(*wallet.Address)
		if err != nil {
			return err
		}

		bf, err := strconv.ParseFloat(balance, 32)
		if err != nil {
			return err
		}
		if float32(bf) < createGame.Stake {
			return errors.New("user not allowed to create game with indicated stake")
		}

		blockIds := []uint64{}
		for _, placement := range createGame.Placements {
			blockIds = append(blockIds, placement.BlockId)
		}

		var blocks []model.Block

		f = tx.Raw("SELECT * FROM block b WHERE b.id IN (?)", blockIds).Scan(blocks)
		if f.Error != nil {
			log.Warn().Msg("error fetching blocks of placements")
			return errors.New("error fetching blocks of placements")
		}
		blockByIds := map[uint64]model.Block{}

		for _, block := range blocks {
			blockByIds[block.Id] = block
		}

		merkle, err := createMerkleTree(createGame.Placements, blockByIds)
		if err != nil {
			return err
		}

		owner, _ := strconv.ParseUint(userId, 10, 64)
		game := &model.Game{
			OwnerId:     owner,
			GameStatus:  model.GameCreated,
			Stake:       uint64(createGame.Stake),
			TimeCreated: time.Now(),
		}
		f = tx.Table("game").Create(game)
		if f.Error != nil {
			log.Warn().Msg("error persisting game to database")
			return f.Error
		}

		userAuthorizer := blockchain.Authorizer{
			KmsResourceId:        wallet.ResourceId,
			ResourceOwnerAddress: *wallet.Address,
		}

		var blockPlacements []*model.BlockPlacement
		for _, placement := range createGame.Placements {
			blockPlacements = append(blockPlacements, &model.BlockPlacement{
				BlockId:     strconv.FormatUint(placement.BlockId, 10),
				UserId:      owner,
				GameId:      game.Id,
				Coordinatex: placement.X,
				Coordinatey: placement.Y,
			})
		}
		f = tx.Table(model.BlockPlacement{}.TableName()).Create(blockPlacements)
		if f.Error != nil {
			log.Warn().Msg("error persisting blocks placements")
			return f.Error
		}

		gs.gameContractBridge.sendCreateGameTx(createGame.Stake, merkle.Root.String(), userAuthorizer)

		return nil
	})

	if err != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	return nil
}

func (gs *gameService) getMoves(gameId uint64, userEmail string) ([]model.MoveHistory, *reject.ProblemWithTrace) {
	var moves []model.MoveHistory
	result := gs.db.
		Model(&model.MoveHistory{}).
		Where("game_id = ? AND user_id = (SELECT id FROM battleblocks_user WHERE email = ?) ORDER BY played_at", gameId, userEmail).
		Find(&moves)

	if result.Error != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	return moves, nil
}

func checkBalance(address string) (string, error) {
	c, err := grpc.NewClient(grpc.TestnetHost)
	if err != nil {
		return "", err
	}
	var adr [8]byte
	copy(adr[:], address)

	balance, err := c.ExecuteScriptAtLatestBlock(context.Background(), []byte(
		`
		import FungibleToken from 0xFUNGIBLE_TOKEN_ADDRESS
		import FlowToken from 0xFLOW_TOKEN_ADDRESS

		pub fun main(account: Address): UFix64 {

		let vaultRef = getAccount(account)
		.getCapability(/public/flowTokenBalance)
		.borrow<&FlowToken.Vault{FungibleToken.Balance}>()
		?? panic("Could not borrow Balance reference to the Vault")

		return vaultRef.balance
		}
		`,
	), []cadence.Value{
		cadence.NewAddress(adr),
	})

	if err != nil {
		return "", err
	}

	return balance.String(), nil
}
