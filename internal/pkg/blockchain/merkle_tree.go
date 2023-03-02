package blockchain

import (
	// "crypto/sha256"
	// "errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/txaty/go-merkletree"

	// mtreeOld "github.com/cbergoon/merkletree"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	// "github.com/wealdtech/go-merkletree"
	keccak "github.com/wealdtech/go-merkletree/keccak256"
)

type TreeContent struct {
	Field []byte
}

func (t *TreeContent) Serialize() ([]byte, error) {
	return t.Field, nil
}


func CreateMerkleTreeNode(x, y int32, present bool, nonce string) *TreeContent {
	// Format: SHIP_PRESENT|X|Y|NONCE
	var sp int8
	if present {
		sp = 1
	}
	log.Printf("%v%v%v%v", sp, x, y, nonce)

	t :=&TreeContent{
		Field: []byte(fmt.Sprintf("%v%v%v%v", sp, x, y, nonce)),
	}

	return t
}

func CreateMerkleTree(presentPlacements []model.Placement, blocksById map[uint64]model.Block) (*merkletree.MerkleTree, []merkletree.DataBlock, error) {
	li := make([][]*TreeContent, 10)
	for i := range li {
		li[i] = make([]*TreeContent, 10)
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			nodeInStr := CreateMerkleTreeNode(int32(i), int32(j), false, randomString())
			li[i][j] = nodeInStr
		}
	}

	for _, placement := range presentPlacements {
		block := blocksById[placement.BlockId]
		ty := block.BlockType
		firstRow := getStringInBetween(ty, "a", "b")
		secondRow := getStringInBetween(ty, "b", "c")

		for _, single := range firstRow {
			singleNr, _ := strconv.ParseUint(string(single), 10, 32)
			nodeInStr := CreateMerkleTreeNode(int32(placement.X)+(int32(singleNr)-1), int32(placement.Y), true, randomString())
			li[int32(placement.X)+int32(singleNr)-1][int32(placement.Y)] = nodeInStr
		}

		for _, single := range secondRow {
			singleNr, _ := strconv.ParseUint(string(single), 10, 32)
			nodeInStr := CreateMerkleTreeNode(int32(placement.X)+int32(singleNr)-1, int32(placement.Y)+1, true, randomString())
			li[int32(placement.X)+int32(singleNr)-1][int32(placement.Y)+1] = nodeInStr
		}
	}

	treeData := []merkletree.DataBlock{}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			treeData = append(treeData, li[i][j])
		}
	}

	conf := &merkletree.Config{
		HashFunc: func(d []byte) ([]byte, error) {
			return keccak.New().Hash(d), nil
		},
		Mode:               merkletree.ModeTreeBuild,
		SortSiblingPairs:   true,
	}

	// mt, err := merkletree.NewUsing(treeData, keccak.New(), nil)
	mt,err := merkletree.New(conf, treeData)
	if err != nil {
		log.Warn().Err(err).Msg("Error while creating merkle tree")
		return nil, nil, err
	}

	return mt, treeData, nil
}

func CreateMerkleTreeFromData(presentData []model.GameGridPoint) (*merkletree.MerkleTree, []merkletree.DataBlock, error) {
	treeData := []merkletree.DataBlock{}
	for _, data := range presentData {
		d := CreateMerkleTreeNode(
			int32(data.CoordinateX),
			int32(data.CoordinateY),
			data.BlockPresent,
			data.Nonce)
		treeData = append(treeData, d)
	}

	// mt, err := merkletree.(treeData, keccak.New(), nil)
	conf := &merkletree.Config{
		HashFunc: func(d []byte) ([]byte, error) {
			return keccak.New().Hash(d), nil
		},
		Mode:               merkletree.ModeProofGenAndTreeBuild,
		SortSiblingPairs:   true,
	}

	// mt, err := merkletree.NewUsing(treeData, keccak.New(), nil)
	mt,err := merkletree.New(conf, treeData)

	if err != nil {
		log.Warn().Err(err).Msg("Error while creating merkle tree")
		return nil, nil, err
	}

	return mt, treeData, nil
}

// TODO remove
/*func CreateMerkleTreeOld(presentPlacements []model.Placement, blocksById map[uint64]model.Block) (*mtreeOld.MerkleTree, error) {
	var li [][]mtreeOld.Content

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			nodeInStr := CreateMerkleTreeNode(int32(i), int32(j), true, randomString(4))
			node := TreeContent{
				Field: nodeInStr,
			}
			li[i][j] = node
		}
	}

	for _, placement := range presentPlacements {
		block := blocksById[placement.BlockId]
		ty := block.BlockType
		firstRow := getStringInBetween(ty, "a", "b")
		secondRow := getStringInBetween(ty, "b", "c")

		for _, single := range firstRow {
			singleNr, _ := strconv.ParseUint(string(single), 10, 32)
			nodeInStr := CreateMerkleTreeNode(int32(placement.X)+(int32(singleNr)-1), int32(placement.Y), true, randomString(4))
			node := TreeContent{
				Field: nodeInStr,
			}
			li[int32(placement.X)+int32(singleNr)][int32(placement.Y)] = node
		}

		for _, single := range secondRow {
			singleNr, _ := strconv.ParseUint(string(single), 10, 32)
			nodeInStr := CreateMerkleTreeNode(int32(placement.X), int32(placement.Y)+(int32(singleNr)-1), true, randomString(4))
			node := TreeContent{
				Field: nodeInStr,
			}
			li[int32(placement.X)][int32(placement.Y)+int32(singleNr)] = node
		}
	}

	flatNrs := []mtreeOld.Content{}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			flatNrs = append(flatNrs, li[i][j])
		}
	}
	mt, _ := mtreeOld.NewTree(flatNrs)

	return mt, nil
}*/

func getStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return ""
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return str[s:]
	}
	return str[s : s+e]
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%05d", rand.Intn(99999-10000)+10000)
}
