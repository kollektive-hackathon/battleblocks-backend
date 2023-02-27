package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cbergoon/merkletree"
	"github.com/kollektive-hackathon/battleblocks-backend/internal/pkg/model"
)

func createMerkleTreeNode(x, y int32, present bool) string {
	// Format: SHIP_PRESENT|X|Y|NONCE
	var sp int8
	if present {
		sp = 1
	}
	return fmt.Sprintf("%v%v%v%v", sp, x, y, randomString(4))
}

func createMerkleTree(presentPlacements []Placement, blocksById map[uint64]model.Block) (*merkletree.MerkleTree, error) {
	var li [][]merkletree.Content

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			nodeInStr := createMerkleTreeNode(int32(i), int32(j), true)
			node := TreeContent{
				field: nodeInStr,
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
			nodeInStr := createMerkleTreeNode(int32(placement.X)+(int32(singleNr)-1), int32(placement.Y), true)
			node := TreeContent{
				field: nodeInStr,
			}
			li[int32(placement.X)+int32(singleNr)][int32(placement.Y)] = node
		}

		for _, single := range secondRow {
			singleNr, _ := strconv.ParseUint(string(single), 10, 32)
			nodeInStr := createMerkleTreeNode(int32(placement.X), int32(placement.Y)+(int32(singleNr)-1), true)
			node := TreeContent{
				field: nodeInStr,
			}
			li[int32(placement.X)][int32(placement.Y)+int32(singleNr)] = node
		}
	}

	flatNrs := []merkletree.Content{}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			flatNrs = append(flatNrs, li[i][j])
		}
	}
	mt, _ := merkletree.NewTree(flatNrs)

	return mt, nil
}

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

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
