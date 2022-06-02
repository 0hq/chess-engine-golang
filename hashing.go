package main

import (
	"fmt"
	"math/rand"

	"github.com/notnil/chess"
)

type hashed struct {
	hash     uint64
	depth    int
	flag     string
	score    int
	best     *chess.Move
	moves    []*chess.Move
	position *chess.Position
}

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

func write_hash(hash uint64, depth int, flag string, score int, best *chess.Move, moves []*chess.Move, position *chess.Position) {
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

func read_hash(hash uint64, depth int, alpha int, beta int) (flag int, score int, best *chess.Move, moves []*chess.Move) {
	p := hash_map[hash]
	if p.flag != "" {
		if p.hash == hash {
			hash_count++
			if p.depth >= depth {
				if p.flag == "EDGE" {
					hash_count_list[0]++
					return 1, p.score, p.best, p.moves
				}
				if p.flag == "ALPHA" && p.score > alpha {
					hash_count_list[1]++
					return 1, p.score, p.best, p.moves
				}
				if p.flag == "BETA" && p.score < beta {
					hash_count_list[2]++
					return 1, p.score, p.best, p.moves
				}
			}
			return 2, 0, p.best, p.moves
		} else {
			fmt.Println("HASH CONFLICT", hash, p.hash, p)
		}
	}
	return 3, 0, nil, nil
}
