package main

import (
	"time"

	"github.com/notnil/chess"
)

/*
Flag
1: Plain
2: Alpha Beta (Move Ordering Bool)
3: Quiescence
4: Hashing

Settings:
Iterative Deepening (sets Default Depth to 1)
Default Depth


*/

const flag int = 4


const DO_MTDF bool = false // forces iterative deepening

const DO_MOVE_ORDERING bool = true
const DO_ITERATIVE_DEEPENING bool = true
const DO_STRICT_TIMING bool = false
const MAX_ITERATIVE_DEPTH int = 12
const TIME_TO_THINK int = 3
const MAX_MOVES = 200
const MAX_QUIESCENCE = -1000

var VERBOSE_FLAG = 3

var DEPTH int = 3       // default value without iterative deepening
const mem_size int = 40 // limits max depth
const MAX_DEPTH int = (mem_size - 1)

var explored int = 0
var hash_count int = 0
var hash_write_count int = 0
var hash_count_list = [3]int{0, 0, 0}
var explored_depth [mem_size]int
var position_eval = 0
var move_count = 1 // just for display
var engine_color = chess.White
var delay time.Time
var opening_moves bool = true // always should be true
var default_start *chess.Move


// var defaultMove chess.Move = chess.Move{chess.Square(1), chess.Square(2)}

// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
// rn1r2k1/ppp3pp/8/2b2b2/4P2q/2P1P3/PP1KQ1BP/RN4NR w - - 0 3
// 1r1r2k1/p4ppp/1bB2q2/5b2/Q7/2P1PN1P/PP3PP1/2KRR3 b - - 0 1
// r1bqkb1r/ppp1ppp1/1Pnp4/4P3/2BP3p/2N2N1P/PP3PP1/R1BQK2R b KQkq - 0 10  // best move is e6, axb6, dxe5 (worse than other 2)
// "r1bq1b1r/1ppkp1p1/1p1p4/3B1Q2/1n1P3p/2N2N1P/PP3PP1/R1B1K2R b KQ - 4 16" simple checkmate
// r1bnkb1r/pp6/3ppNp1/4P3/2BP3p/5N1P/PP3PP1/R1BQK2R b KQkq - 0 14"
// r1b1kbnr/pp2pppp/2P2q2/8/8/2N5/PPP1BPPP/R1BQK1NR b KQkq - 0 8
// r1b1kbnr/pp2pppp/2n1q3/3P4/8/2N5/PPP1BPPP/R1BQK1NR b KQkq - 0 7
// r3kb1r/1b2pppp/p3q3/3N4/8/5B2/PPP2PPP/R2Q1RK1 w kq - 6 12
// r3kb1r/pb2pppp/5q2/8/8/2N5/PPP1BPPP/R2QK2R b KQkq - 0 8
// 3k1b1r/4pppp/p1Q5/8/1q6/5B2/PPP2PPP/3R1RK1 b - - 3 1
// 8/1Kn1p3/1p5N/4p1q1/4k1N1/3R2p1/Qn2B3/7R w - - 0 1 // very hard checkmate in 3
// r3kb1r/1b2pppp/pq6/3N4/8/5B2/PPP2PPP/R2Q1RK1 b kq - 5 11
// r3kb1r/1b2pppp/pq6/8/8/2N2B2/PPP2PPP/R2Q1RK1 b kq - 5 11 good white +2 midgame 
// r1bqkb1r/ppp1ppp1/1Pnp4/4P3/2BP3p/2N2N1P/PP3PP1/R1BQK2R w KQkq - 0 10 blunder
// "r2qkb2/1pp1p1p1/1pn2p2/4pb2/2BP3r/2N2N1P/PP1K1PP1/R2Q3R w q - 0 15" // f3h4 d8d4 c4d3 d4d3
// quiescence testing good 4k3/5q2/1p6/1Pp5/2Pp4/3Pp3/4P3/1K3R2 w - - 0 1
// quiescence testing bad 4k3/8/1p6/1Pp5/2Pp4/3Pp3/4P3/1K3Q2 w - - 0 1
// 3rq1k1/4br1p/2ppb1p1/p5Pn/N3P2P/4Q3/1PP3BB/1K1RR3 w - - 1 22
// rnbqkbnr/ppp2ppp/8/3pp3/8/6PP/PPPPPP2/RNBQKBNR w KQkq d6 0 3
var start_pos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"


// Max returns the larger of x or y.
func Max(x, y int) int {
    if x < y {
        return y
    }
    return x
}


// code storage


// if len(game.ValidMoves()) != len(moves) {
// 	fmt.Println(game.Position().Board().Draw())
// 	fmt.Println("\ndepth:", index_depth, depth, game.Position())
// 	fmt.Println(game.ValidMoves(), len(game.ValidMoves()))
// 	fmt.Println(moves, len(moves), move_order(game, game.ValidMoves()))
// 	fmt.Println("\n", flag, hashscore, hashbest, depthfound)
// 	fmt.Println(hash_map[zobrist(game.Position().Board(), max)])
// 	fmt.Println(hash_map[zobrist(game.Position().Board(), max)].position)
// 	panic("AHHHHHH")
// }
