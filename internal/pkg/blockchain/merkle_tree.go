package blockchain

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"math/rand"
	"strconv"
	"strings"

	mtreeOld "github.com/cbergoon/merkletree"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
	merkletree "github.com/wealdtech/go-merkletree"
)

type TreeContent struct {
	Field string
}

//CalculateHash hashes the values of a TestContent
func (t TreeContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.Field)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

//Equals tests for equality of two Contents
func (t TreeContent) Equals(other mtreeOld.Content) (bool, error) {
	otherTC, ok := other.(TreeContent)
	if !ok {
		return false, errors.New("value is not of type TestContent")
	}
	return t.Field == otherTC.Field, nil
}

func CreateMerkleTreeNode(x, y int32, present bool, nonce string) string {
	// Format: SHIP_PRESENT|X|Y|NONCE
	var sp int8
	if present {
		sp = 1
	}
	return fmt.Sprintf("%v%v%v%v", sp, x, y, nonce)
}

func CreateMerkleTree(presentPlacements []model.Placement, blocksById map[uint64]model.Block) (*merkletree.MerkleTree, error) {
	var li [][]string

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			nodeInStr := CreateMerkleTreeNode(int32(i), int32(j), true, randomString())
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
			li[int32(placement.X)+int32(singleNr)][int32(placement.Y)] = nodeInStr
		}

		for _, single := range secondRow {
			singleNr, _ := strconv.ParseUint(string(single), 10, 32)
			nodeInStr := CreateMerkleTreeNode(int32(placement.X), int32(placement.Y)+(int32(singleNr)-1), true, randomString())
			li[int32(placement.X)][int32(placement.Y)+int32(singleNr)] = nodeInStr
		}
	}

	treeData := [][]byte{}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			treeData = append(treeData, []byte(li[i][j]))
		}
	}

	mt, err := merkletree.New(treeData)
	if err != nil {
		log.Warn().Err(err).Msg("Error while creating merkle tree")
		return nil, err
	}

	return mt, nil
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
	return fmt.Sprintf("%05d",rand.Intn(10000))
}
