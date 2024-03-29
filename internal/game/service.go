package game

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	// merkletree "github.com/wealdtech/go-merkletree"

	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/blockchain"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/reject"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/utils"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/rs/zerolog/log"
	// keccak "github.com/wealdtech/go-merkletree/keccak256"
	"gorm.io/gorm"
)

const (
	databaseError = "error.data.access"
)

type gameService struct {
	db                 *gorm.DB
	gameContractBridge *gameContractBridge
}

type GameResponse struct {
	model.Game
	OwnerName      *string `json:"ownerUsername"`
	ChallengerName *string `json:"challengerUsername"`
}

type MoveHistoryWithHit struct {
	ID          uint64 `gorm:"column:id" json:"id"`
	UserID      uint64 `gorm:"column:user_id" json:"userId"`
	GameID      uint64 `gorm:"column:game_id" json:"gameId"`
	Coordinatex int    `gorm:"column:coordinatex" json:"x"`
	Coordinatey int    `gorm:"column:coordinatey" json:"y"`
	PlayedAt    uint64 `gorm:"column:played_at" json:"playedAt"`
	IsHit       bool   `gorm:"-" json:"isHit"`
}

func (gs *gameService) getGames(page utils.PageRequest, userEmail string) ([]GameResponse, *int64, *reject.ProblemWithTrace) {
	games := []GameResponse{}
	gamesSize := int64(0)

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var userId string
		f := tx.Raw("SELECT u.id FROM battleblocks_user u WHERE email = ?", userEmail).First(&userId)
		if f.Error != nil {
			return f.Error
		}

		res := tx.Table("game").
			Where("game.game_status IN ('CREATED', 'PLAYING')").
			Where("(game.owner_id = ? OR game.challenger_id = ? OR game.challenger_id IS NULL)", userId, userId).
			Count(&gamesSize)
		if res.Error != nil {
			return res.Error
		}

		res = tx.Raw(`
			SELECT game.*, owner.username AS owner_name, challenger.username AS challenger_name FROM game
			JOIN battleblocks_user AS owner ON game.owner_id = owner.id
			LEFT JOIN battleblocks_user AS challenger ON game.challenger_id = challenger.id
			WHERE game.game_status IN ('CREATED', 'PLAYING') AND
			(game.owner_id = $1 OR game.challenger_id = $1 OR game.challenger_id IS NULL)
			ORDER BY
			(owner_id = $1 AND game_status = 'PLAYING') DESC,(owner_id = $1) DESC, (challenger_id = $1) DESC, (game_status = 'PLAYING') DESC, time_created DESC
			LIMIT $2
			OFFSET $3`, userId, page.Size, page.Offset).Scan(&games)

		// tx.Table("game").Joins("JOIN battleblocks_user AS owner ON game.owner_id = owner.id").
		// Joins("LEFT JOIN battleblocks_user AS challenger ON game.challenger_id = challenger.id").
		// Select("game.*, owner.username AS owner_name, challenger.username AS challenger_name").
		// Where("game.game_status IN ('CREATED', 'PLAYING')").
		// Where("(game.owner_id = ? OR game.challenger_id = ? OR game.challenger_id IS NULL)", userId, userId).
		// Limit(page.Size).
		// Offset(page.Offset).
		// Clauses(clause.OrderBy{
		// Expression: clause.Expr{
		// SQL:                "(owner_id = $1 AND game_status = 'PLAYING') DESC,(owner_id = $1) DESC, (challenger_id = $1) DESC, (game_status = 'PLAYING') DESC, time_created DESC",
		// Vars:               []interface{}{userId},
		// WithoutParentheses: true,
		// },
		// }).
		// Scan(&games)

		// res = tx.Table("game").
		// Limit(page.Size).
		// Offset(page.Offset).
		// Clauses(clause.OrderBy{
		// Expression: clause.Expr{
		// SQL:                "(owner_id = $1 AND game_status = 'PLAYING') DESC, (owner_id = $1) DESC, time_created DESC",
		// Vars:               []interface{}{userId},
		// WithoutParentheses: true,
		// },
		// }).
		// Scan(&games)
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

func (gs *gameService) joinGame(joinGame JoinGameRequest, gameId uint64, userEmail string) *reject.ProblemWithTrace {
	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var userId string
		f := tx.Raw("SELECT u.id FROM battleblocks_user u WHERE email = ?", userEmail).First(&userId)
		if f.Error != nil {
			return f.Error
		}

		var game model.Game
		f = tx.Raw("SELECT * FROM game u WHERE id = ?", gameId).First(&game)
		if f.Error != nil {
			return f.Error
		}

		wallet := gs.getCustodialWallet(userEmail)
		if wallet == nil {
			return errors.New("wallet does not exist")
		}

		balance, err := checkBalance(*wallet.Address)
		if err != nil {
			return err
		}

		bf, err := strconv.ParseFloat(balance, 32)
		if err != nil {
			return err
		}
		if float32(bf) < (float32(game.Stake) + 1) {
			return errors.New("user not allowed to create game with indicated stake")
		}

		blockIds := []uint64{}
		for _, placement := range joinGame.Placements {
			blockIds = append(blockIds, placement.BlockId)
		}

		var blocks []model.Block

		f = tx.Raw("SELECT * FROM block b WHERE b.id IN (?)", blockIds).Scan(&blocks)
		if f.Error != nil {
			log.Warn().Msg("error fetching blocks of placements")
			return f.Error
		}

		blockByIds := map[uint64]model.Block{}
		for _, block := range blocks {
			blockByIds[block.Id] = block
		}

		// mtree , _ , _ := blockchain.CreateMerkleTree(joinGame.Placements, blockByIds)
		merkle, mtreeData, err := blockchain.CreateMerkleTree(joinGame.Placements, blockByIds)
		if err != nil {
			return err
		}

		owner, _ := strconv.ParseUint(userId, 10, 64)
		var points []*model.GameGridPoint
		for _, singlePoint := range mtreeData {
			sp, _ := singlePoint.Serialize()
			points = append(points, pointFromData(string(sp), gameId, owner))
		}

		f = tx.Table("game_grid_point").Create(points)
		if f.Error != nil {
			log.Warn().Msg("cannot create game grid points")
			return f.Error
		}

		userAuthorizer := blockchain.Authorizer{
			KmsResourceId:        wallet.ResourceId,
			ResourceOwnerAddress: *wallet.Address,
		}

		var blockPlacements []*model.BlockPlacement
		for _, placement := range joinGame.Placements {
			blockPlacements = append(blockPlacements, &model.BlockPlacement{
				BlockId:     strconv.FormatUint(placement.BlockId, 10),
				UserId:      owner,
				GameId:      game.Id,
				Coordinatex: placement.X,
				Coordinatey: placement.Y,
			})
		}

		f = tx.Table(model.BlockPlacement{}.TableName()).Create(&blockPlacements)
		if f.Error != nil {
			log.Warn().Msg("error persisting blocks placements")
			return f.Error
		}

		gs.gameContractBridge.sendJoinGame(float32(game.Stake), merkle.Root, *game.FlowId, userAuthorizer)
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

func (gs *gameService) createGame(createGame CreateGameRequest, userEmail string) (*model.Game, *reject.ProblemWithTrace) {
	var createdGame *model.Game
	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var userId string
		f := tx.Raw("SELECT u.id FROM battleblocks_user u WHERE email = ?", userEmail).First(&userId)
		if f.Error != nil {
			return f.Error
		}

		var wallet model.CustodialWallet
		f = tx.Raw(`SELECT cw.* FROM battleblocks_user bu
			LEFT JOIN custodial_wallet cw ON bu.custodial_wallet_id = cw.id 
			WHERE bu.id = ?`, userId).
			First(&wallet)

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
		if float32(bf) < (createGame.Stake + 1) {
			return errors.New("user not allowed to create game with indicated stake")
		}

		blockIds := []uint64{}
		for _, placement := range createGame.Placements {
			blockIds = append(blockIds, placement.BlockId)
		}

		var blocks []model.Block

		f = tx.Raw("SELECT * FROM block b WHERE b.id IN (?)", blockIds).Scan(&blocks)
		if f.Error != nil {
			log.Warn().Msg("error fetching blocks of placements")
			return errors.New("error fetching blocks of placements")
		}
		blockByIds := map[uint64]model.Block{}

		for _, block := range blocks {
			blockByIds[block.Id] = block
		}

		merkle, mtreeData, err := blockchain.CreateMerkleTree(createGame.Placements, blockByIds)
		if err != nil {
			return err
		}

		owner, _ := strconv.ParseUint(userId, 10, 64)
		createdGame = &model.Game{
			OwnerId:     owner,
			GameStatus:  model.GamePreparing,
			Stake:       uint64(createGame.Stake),
			TimeCreated: time.Now().UTC().UnixMilli(),
		}
		f = tx.Table("game").Create(&createdGame)
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
				GameId:      createdGame.Id,
				Coordinatex: placement.X,
				Coordinatey: placement.Y,
			})
		}
		f = tx.Table(model.BlockPlacement{}.TableName()).Create(&blockPlacements)
		if f.Error != nil {
			log.Warn().Msg("error persisting blocks placements")
			return f.Error
		}

		var points []*model.GameGridPoint
		for _, singlePoint := range mtreeData {
			sp, _ := singlePoint.Serialize()
			points = append(points, pointFromData(string(sp), createdGame.Id, owner))
		}

		f = tx.Table("game_grid_point").Create(&points)
		if f.Error != nil {
			log.Warn().Msg("cannot create game grid points")
			return f.Error
		}

		gs.gameContractBridge.sendCreateGameTx(createGame.Stake, merkle.Root, createdGame.Id, userAuthorizer)

		return nil
	})

	if err != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(err),
			Cause:   err,
		}
	}

	return createdGame, nil
}

type PlacementsView struct {
	ColorHex  string `json:"colorHex"`
	Pattern   string `json:"pattern"`
	X         string `json:"x"`
	Y         string `json:"y"`
	BlockType string `json:"blockType"`
}

func (gs *gameService) getPlacements(gameId uint64, userEmail string) ([]PlacementsView, *reject.ProblemWithTrace) {
	var placements []PlacementsView

	result := gs.db.Raw(`SELECT b.color_hex as color_hex, b.pattern as pattern, b.block_type as block_type, bp.coordinatex as X, bp.coordinatey as Y 
		FROM block_placement bp
		JOIN block b on bp.block_id = b.id
		WHERE game_id = ? AND user_id =
		(SELECT id FROM battleblocks_user WHERE email = ?)`, gameId, userEmail).Find(&placements)

	if result.Error != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}
	return placements, nil
}

func (gs *gameService) getMoves(gameID uint64, userEmail string) ([]MoveHistoryWithHit, *reject.ProblemWithTrace) {
	moves := []MoveHistoryWithHit{}
	result := gs.db.Table("move_history").
		Where("game_id = ?", gameID).
		Order("played_at DESC").
		Find(&moves)

	if result.Error != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	for i := range moves {
		var isHit bool
		rslt := gs.db.Table("game_grid_point").
			Where("game_id = ? AND coordinate_x = ? AND coordinate_y = ? AND block_present = true",
				moves[i].GameID, moves[i].Coordinatex, moves[i].Coordinatey).
			Select("EXISTS(SELECT 1 FROM game_grid_point WHERE game_id = ? AND coordinate_x = ? AND coordinate_y = ? AND block_present = true)",
				moves[i].GameID, moves[i].Coordinatex, moves[i].Coordinatey).
			Scan(&isHit)

		if rslt.Error != nil {
			log.Warn().Err(result.Error).Msg("Cannot fetch isHit for player move")
			// should have proper ws error signal implemented
			// but not necessary for this poc
			isHit = false
		}

		moves[i].IsHit = isHit
	}

	return moves, nil
}

func (gs *gameService) getGame(gameId uint64) (*GameResponse, *reject.ProblemWithTrace) {
	game := GameResponse{}

	result := gs.db.
		Table("game").
		Joins("JOIN battleblocks_user AS owner ON game.owner_id = owner.id").
		Joins("LEFT JOIN battleblocks_user AS challenger ON game.challenger_id = challenger.id").
		Select("game.*, owner.username AS owner_name, challenger.username AS challenger_name").
		Where("game.id = ?", gameId).
		First(&game)

	if result.Error != nil {
		return nil, &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	return &game, nil
}

func (gs *gameService) playMove(gameId uint64, userEmail string, request PlayMoveRequest) *reject.ProblemWithTrace {
	var currentUserData []model.GameGridPoint

	result := gs.db.
		Table("game_grid_point").
		Where("game_id = ? AND user_id = (SELECT bu.id FROM battleblocks_user bu WHERE email = ?)", gameId, userEmail).
		Find(&currentUserData)

	if result.Error != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	mtree, _, err := blockchain.CreateMerkleTreeFromData(currentUserData)

	if err != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}
	var user model.User
	result = gs.db.
		Model(&model.User{}).
		Where("email = ?", userEmail).
		Find(&user)

	if result.Error != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	var game model.Game
	result = gs.db.
		Model(&model.Game{}).
		Where("id = ?", gameId).
		Find(&game)
	if result.Error != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(result.Error),
			Cause:   result.Error,
		}
	}

	opponent := *game.ChallengerId
	if *game.ChallengerId == user.Id {
		opponent = game.OwnerId
	}

	isFirstMove := gs.isFirstMove(gameId)
	if isFirstMove {
		cw := gs.getCustodialWallet(userEmail)
		if cw == nil {
			walletNotExistsErr := fmt.Errorf("custodial wallet not found while making move, user email %s", userEmail)
			return &reject.ProblemWithTrace{
				Problem: reject.UnexpectedProblem(walletNotExistsErr),
				Cause:   walletNotExistsErr,
			}
		}

		userAuthorizer := blockchain.Authorizer{KmsResourceId: cw.ResourceId, ResourceOwnerAddress: *cw.Address}

		fProof := [][]uint8{{}}

		gs.gameContractBridge.sendMove(*game.FlowId, request.X, request.Y, fProof,
			nil, nil, nil, nil, userAuthorizer)
		return nil

	}
	opponentProofData, proofDataLoadErr := gs.getLastOpponentMoveProofData(gameId, opponent, user.Id)

	if proofDataLoadErr != nil {
		return &reject.ProblemWithTrace{
			Problem: reject.UnexpectedProblem(proofDataLoadErr),
			Cause:   proofDataLoadErr,
		}
	}

	if proofDataLoadErr != nil {
		log.Info().Msg("Could not fetch last opponent data")
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

	proof, err := mtree.Proof(proofNode)
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

	nonceNumber, _ := strconv.ParseUint(opponentProofData.Nonce, 10, 64)

	// verify, err := merkletree.VerifyProofUsing([]byte(blockchain.CreateMerkleTreeNode(int32(opponentProofData.CoordinateX), int32(opponentProofData.CoordinateY), opponentProofData.BlockPresent, fmt.Sprint(nonceNumber))), proof, mtree.Root(), keccak.New(), nil)

	// verifyProof, err := merkletree.VerifyProofUsing([]byte(proofNode), proof, mtree.Root(), keccak.New(), nil)

	log.Error().Interface("nonce", nonceNumber).Msg("Nonce number")

	// log.Error().Interface("verify root", verify).Msg("VERIFY ROOT DEBUG:")

	// log.Error().Interface("verify proof", verifyProof).Msg("VERIFY PROOF DEBUG:")

	log.Error().Interface("proof", proof).Msg("LOG PROOF:")

	// log.Error().Interface("proof hashes", proof.).Msg("LOG PROOF:")

	gs.gameContractBridge.sendMove(*game.FlowId, request.X, request.Y, proof.Siblings,
		&opponentProofData.BlockPresent, &opponentProofData.CoordinateX, &opponentProofData.CoordinateY, &nonceNumber, userAuthorizer)

	return nil
}

func (gs *gameService) isFirstMove(gameId uint64) bool {
	var moves []model.MoveHistory
	gs.db.Raw("SELECT * FROM move_history mh WHERE mh.game_id = ?", gameId).Scan(&moves)
	if len(moves) > 0 {
		return false
	}
	return true
}

func (gs *gameService) getLastOpponentMoveProofData(gameId uint64, opponentId uint64, currUserId uint64) (*model.GameGridPoint, error) {
	var mh model.MoveHistory

	result := gs.db.Raw(`select move_history.* from
		move_history WHERE move_history.game_id = ? AND move_history.user_id = ? ORDER BY played_at DESC LIMIT 1`, gameId, opponentId).Scan(&mh)
	if result.Error != nil {
		return nil, result.Error
	}

	var proofData model.GameGridPoint
	result = gs.db.Raw(`
		SELECT ggp.coordinate_x as coordinatex, ggp.coordinate_y as coordinatey,ggp.* FROM game_grid_point ggp where ggp.coordinate_x = ? AND coordinate_y = ? AND ggp.game_id = ? AND ggp.user_id = ?
		`, mh.Coordinatex, mh.Coordinatey, gameId, currUserId).First(&proofData)

	if result.Error != nil {
		return nil, result.Error
	}

	log.Error().Interface("proof data", proofData).Msg("PROOF DATA:")
	// result := gs.db.Raw(`
	// select game_grid_point.* from game_grid_point
	// LEFT JOIN move_history mh ON game_grid_point.game_id = mh.game_id
	// AND mh.coordinatey = game_grid_point.coordinate_y
	// AND mh.coordinatex = game_grid_point.coordinate_x
	// AND mh.user_id = $2
	// WHERE game_grid_point.user_id = $1 AND game_grid_point.game_id = $3 order by mh.played_at ASC LIMIT 1
	// `, currUserId, opponentId, gameId).First(&proofData)

	// if result.Error != nil {
	// return nil, result.Error
	// }

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
	txCode := `
		import FungibleToken from 0xFUNGIBLE_TOKEN_ADDRESS
		import FlowToken from 0xFLOW_TOKEN_ADDRESS

		pub fun main(account: Address): UFix64 {

		let vaultRef = getAccount(account)
		.getCapability(/public/flowTokenBalance)
		.borrow<&FlowToken.Vault{FungibleToken.Balance}>()
		?? panic("Could not borrow Balance reference to the Vault")

		return vaultRef.balance
		}
		`

	addressTemplates := map[string]string{
		"0xFLOW_TOKEN_ADDRESS":     viper.Get("FLOW_TOKEN_ADDRESS").(string),
		"0xFUNGIBLE_TOKEN_ADDRESS": viper.Get("FUNGIBLE_TOKEN_ADDRESS").(string),
	}

	for k, v := range addressTemplates {
		txCode = strings.ReplaceAll(txCode, k, v)
	}

	c, err := grpc.NewClient(grpc.TestnetHost)
	if err != nil {
		return "", err
	}
	var adr [8]byte
	copy(adr[:], address)

	flowAddress := flow.HexToAddress(address)

	cadenceAddress := cadence.BytesToAddress(flowAddress.Bytes())

	args := []cadence.Value{cadence.Address(cadenceAddress)}

	balance, err := c.ExecuteScriptAtLatestBlock(context.Background(), []byte(
		txCode,
	), args)

	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

func pointFromData(singlePoint string, gameId uint64, userId uint64) *model.GameGridPoint {
	p := singlePoint[:1]
	x := singlePoint[1:2]
	y := singlePoint[2:3]
	nonce := singlePoint[3:8]
	var present bool
	if p == "1" {
		present = true
	}
	cordX, _ := strconv.ParseUint(x, 10, 64)
	cordY, _ := strconv.ParseUint(y, 10, 64)

	return &model.GameGridPoint{
		GameId:       gameId,
		UserId:       userId,
		BlockPresent: present,
		CoordinateX:  cordX,
		CoordinateY:  cordY,
		Nonce:        nonce,
	}
}
