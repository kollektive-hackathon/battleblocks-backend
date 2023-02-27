package game

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

type gameService struct {
	db                 *gorm.DB
	gameContractBridge *gameContractBridge
}

func (gs *gameService) getGames(page utils.PageRequest, userEmail string) ([]model.Game, *int64, *reject.ProblemWithTrace) {
	games := []model.Game{}
	gamesSize := int64(0)

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var userId string
		f := tx.Raw("SELECT u.id FROM battleblocks_user u WHERE email = ?", userEmail).First(&userId)
		if f.Error != nil {
			return f.Error
		}

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

func (gs *gameService) createGame(createGame CreateGameRequest, userEmail string) *reject.ProblemWithTrace {
	// TODO Send tx for creating game with the model
	// Spremiti u bazu game sa prizeom koji prima

	//placements also sent here
	//TODO: CHECK balance ---- execute script  with golang-flow-sdk
	// create merkeltree, send tx - with the ROOT and stake.

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var userId string
		f := tx.Raw("SELECT u.id FROM battleblocks_user u WHERE email = ?", userEmail).First(&userId)
		if f.Error != nil {
			return f.Error
		}

		var wallet model.CustodialWallet
		f = tx.Raw(`SELECT cw FROM battleblocks_user bu
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

		merkle, err := blockchain.CreateMerkleTree(createGame.Placements, blockByIds)
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

		gs.gameContractBridge.sendCreateGameTx(createGame.Stake, string(merkle.Root()), userAuthorizer)

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

func (gs *gameService) playMove(gameId uint64, userEmail string, request PlayMoveRequest) *reject.ProblemWithTrace {
	var currentUserBlockPlacements []model.Placement
	result := gs.db.
		Model(&model.BlockPlacement{}).
		Where("game_id = ? AND user_id = ?").
		Select("block_id, coordinate_x AS x coordinate_y AS y").
		Find(&currentUserBlockPlacements)

	if result.Error != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	blockIds := []uint64{}
	for _, v := range currentUserBlockPlacements {
		blockIds = append(blockIds, v.BlockId)
	}

	var blocks []model.Block
	result = gs.db.
		Model(&model.Block{}).
		Where("id IN ?", blockIds).
		Find(&blocks)

	if result.Error != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	blockMap := map[uint64]model.Block{}
	for _, v := range blocks {
		blockMap[v.Id] = v
	}

	mtree, err := blockchain.CreateMerkleTree(currentUserBlockPlacements, blockMap)
	if err != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	opponentProofData, proofDataLoadErr := gs.getLastOpponentMoveProofData(gameId, userEmail)
	if proofDataLoadErr != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	proofNode := blockchain.CreateMerkleTreeNode(
		int32(opponentProofData.CoordinateX),
		int32(opponentProofData.CoordinateY),
		opponentProofData.BlockPresent,
		opponentProofData.Nonce)

	proof, err := mtree.GenerateProof([]byte(proofNode))
	if err != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	cw := gs.getCustodialWallet(userEmail)
	if cw == nil {
		walletNotExistsErr := fmt.Errorf("custodial wallet not found while making move, user email %s", userEmail)
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(walletNotExistsErr),
			Cause:   walletNotExistsErr,
		}
	}

	userAuthorizer := blockchain.Authorizer{KmsResourceId: cw.ResourceId, ResourceOwnerAddress: *cw.Address}
	if opponentProofData == nil {
		gs.gameContractBridge.sendMove(gameId, request.X, request.Y, nil,
			nil, nil, nil, nil, userAuthorizer)
	} else {
		nonceNumber, _ := strconv.ParseUint(opponentProofData.Nonce, 0, 64)
		gs.gameContractBridge.sendMove(gameId, request.X, request.Y, &proof.Hashes,
			&opponentProofData.BlockPresent, &opponentProofData.CoordinateX, &opponentProofData.CoordinateY, &nonceNumber, userAuthorizer)
	}

	return nil
}

func (gs *gameService) getLastOpponentMoveProofData(gameId uint64, userEmail string) (*model.GameGridPoint, *reject.ProblemWithTrace) {
	var proofData model.GameGridPoint
	result := gs.db.Raw(`
		SELECT game_grid_point.game_id
             , game_grid_point.user_id
             , game_grid_point.block_present
             , game_grid_point.coordinate_x
             , game_grid_point.coordinate_y
             , game_grid_point.nonce
          FROM game_grid_point
	INNER JOIN move_history 
			ON move_history.game_id = game_grid_point.game_id 
		   AND move_history.user_id = game_grid_point.user_id
         WHERE move_history.game_id = ?
           AND move_history.user_id = (SELECT id FROM battleblocks_user WHERE email = ?)
	ORDER BY played_at DESC LIMIT 1
    `, gameId, userEmail).Scan(&proofData)

	if result.Error != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	return &proofData, nil
}

func (gs *gameService) getCustodialWallet(userEmail string) *model.CustodialWallet {
	var custodialWallet model.CustodialWallet
	result := gs.db.
		Model(&custodialWallet).
		Where("id = (SELECT custodial_wallet_id FROM battleblocks_user WHERE email = ?)", userEmail).
		First(&custodialWallet)

	if result.Error != nil {
		return nil
	}

	return &custodialWallet
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
