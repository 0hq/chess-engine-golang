package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/notnil/chess"
)

type hashed struct {
	hash     uint64
	depth    int
	flag     HashFlag
	score    int
	best     *chess.Move
	moves    []*chess.Move
	position *chess.Position
}

// PieceType is the type of a piece.
type HashFlag int8

const (
	// Null flag
	NoFlag HashFlag = iota
	// Edge of the search
	EdgeFlag
	// Non-edge from a maximizer 
	AlphaFlag
	// Non-edge from a minimizer
	BetaFlag
	// Non-edge from a maximizer in quiescence
	AlphaQFlag
	// Non-edge from a minimizer in quiescence
	BetaQFlag
)

// PieceType is the type of a piece.
type HashResult int8

const (
	// Null result
	NoResult HashResult = iota
	// Result from deeper edge search
	DeeperResult
	// Result from deeper quiescence search
	QuiescenceDeeperResult
	// Moves saved from previous Minimax search
	MinimaxSavedMoves
	// Moves saved from previous quiescence search
	QuiescenceSavedMoves	
)

var hash_map = make(map[uint64]hashed)
var whiteToMoveZobrist uint64
var pieceSquareZobrist [12][64]uint64
var castleRightsZobrist [4]uint64

func generateZobristConstants() {
	whiteToMoveZobrist = rand.Uint64()
	for i := 0; i < 12; i++ {
		for j := 0; j < 64; j++ {
			pieceSquareZobrist[i][j] = rand.Uint64()
		}
	}
	for i := 0; i < 4; i++ {
		castleRightsZobrist[i] = rand.Uint64()
	}
}

func zobrist(board *chess.Board, max bool) uint64 {
	var bits uint64 = 0
	pos := board.SquareMap()
	if max {
		bits = bits ^ whiteToMoveZobrist
	}
	for square, piece := range pos {
		value := pieceSquareZobrist[int(piece.Type())-1][int(square)]
		bits = bits ^ value
	}
	return bits
}

func write_hash(hash uint64, depth int, flag HashFlag, score int, best *chess.Move, moves []*chess.Move, position *chess.Position) {
	hash_write_count++
	p := hashed{
		hash:     hash,
		depth:    depth,
		flag:     flag,
		score:    score,
		best:     best,
		moves:    moves,
		position: position,
	}
	hash_map[hash] = p
}

func read_hash(hash uint64, depth int, alpha int, beta int) (flag HashResult, score int, best *chess.Move, moves []*chess.Move, depthfound int) {
	p := hash_map[hash]
	if p.flag != 0 {
		if p.hash == hash {
			hash_count++
			if p.depth >= depth {
				if p.flag == EdgeFlag {
					hash_count_list[0]++
					return DeeperResult, p.score, nil, nil, p.depth
				}
				if p.flag == AlphaFlag && p.score > alpha {
					hash_count_list[1]++
					return DeeperResult, p.score, p.best, p.moves, p.depth
				}
				if p.flag == BetaFlag && p.score < beta {
					hash_count_list[2]++
					return DeeperResult, p.score, p.best, p.moves, p.depth
				}
				if p.flag == AlphaQFlag {
					hash_count_list[0]++
					return QuiescenceDeeperResult, p.score, p.best, p.moves, p.depth
				}
				if p.flag == BetaQFlag {
					hash_count_list[0]++
					return QuiescenceDeeperResult, p.score, p.best, p.moves, p.depth
				}
			}
			// this has no data but make sure it doesn't get sent
			if p.flag == EdgeFlag {
				return NoResult, int(math.NaN()), nil, nil, 0
			} else if p.flag == BetaQFlag || p.flag == AlphaQFlag {
				return QuiescenceSavedMoves, int(math.NaN()), p.best, p.moves, p.depth
			} else {
				return MinimaxSavedMoves, int(math.NaN()), p.best, p.moves, p.depth
			}
		} else {
			fmt.Println("HASH CONFLICT", hash, p.hash, p)
		}
	}
	return NoResult, int(math.NaN()), nil, nil, 0
}
