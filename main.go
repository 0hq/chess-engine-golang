package main

import (
	"fmt"
	"math"
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

const DO_MOVE_ORDERING bool = true

var DO_ITERATIVE_DEEPENING bool = false
var TIME_TO_THINK int = 2
var DEPTH int = 2 // default value without iterative deepening
var MAX_MOVES = 10000
var MAX_QUIESCENCE = -10

var explored int = 0
var hash_count int = 0
var hash_count_list = [3]int{0, 0, 0}
var explored_depth [50]int
var position_eval = 0
var move_count = 0
var engine_color = chess.Black

// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
// rn1r2k1/ppp3pp/8/2b2b2/4P2q/2P1P3/PP1KQ1BP/RN4NR w - - 0 3
var start_pos = "1r1r2k1/p4ppp/1bB2q2/5b2/Q7/2P1PN1P/PP3PP1/2KRR3 b - - 0 1"

func main() {
	game := setup()
	eng := init_stockfish()

	for game.Outcome() == chess.NoOutcome && move_count < MAX_MOVES {
		var move *chess.Move
		color := game.Position().Turn()
		start := time.Now()

		fmt.Println("\n\n---- New Turn ----", color)

		if color == engine_color {
			move = engine(game)
		} else {
			// move = random_move_engine()
			move = stockfish(game, eng)
		}

		pre := game.Clone()
		game.Move(move)

		print_turn_complete(game, move, start)
		update_evaluation(game, pre, move)
		move_count++
	}
	print_game_over(game)
}

func setup() *chess.Game {
	fmt.Println("\n\nStart game...")
	init_explored_depth()
	init_hash_count()
	generateZobristConstants()
	fen, _ := chess.FEN(start_pos)
	return chess.NewGame(fen)
}

func engine(game *chess.Game) (output *chess.Move) {
	explored = 0
	init_explored_depth()
	if DO_ITERATIVE_DEEPENING {
		output = iterative_deepening(game, TIME_TO_THINK)
	} else {
		output, _ = minimax_factory(game, 0)
	}
	return
}

func iterative_deepening(game *chess.Game, time_control int) (output *chess.Move) {
	delay := time.Now()
	delay = delay.Add(time.Second * time.Duration(time_control))
	DEPTH = 1 // starting depth

	var total_hash, total_explored int
	var total_hash_list [3]int
	var total_explored_list [50]int

	for time.Now().Sub(delay) < 0 {
		print_iter_1(delay)
		output, _ = minimax_factory(game, 0)
		DEPTH++
		print_iter_2(output, 0)

		deepening_counts(&total_hash, &total_explored, &total_hash_list, &total_explored_list)
	}
	return
}

// this should be static, not relative
func update_evaluation(game *chess.Game, pre *chess.Game, move *chess.Move) {
	position_eval = evaluate_position(pre, game, position_eval, move)
}

// ----- print statements to clean up code ----

func print_iter_1(delay time.Time) {
	fmt.Println("new interation with depth: ", DEPTH, "time left:", delay.Sub(time.Now()))
}

func print_iter_2(output *chess.Move, eval int) {
	fmt.Println("\n", "done", output, eval, "\n")
	fmt.Println(hash_count)
	fmt.Println(hash_count_list)
	fmt.Println(explored)
	fmt.Println(explored_depth)
}

func print_turn_complete(game *chess.Game, move *chess.Move, start time.Time) {
	fmt.Println("\nMove chosen:", move)
	end := time.Now()
	fmt.Println(game.Position().Board().Draw())
	fmt.Println("Time elapsed", end.Sub(start))
	fmt.Println(game.Position())
}

func print_game_over(game *chess.Game) {
	fmt.Printf("\n\n ----- Game completed. %s by %s.\n ------", game.Outcome(), game.Method())
	fmt.Println(game)
	fmt.Println(game.Position())
}

// ------ now entering the doldrums -----

func minimax_factory(game *chess.Game, preval int) (best *chess.Move, eval int) {
	if flag == 4 {
		best, eval, _ = minimax_hashing(game, DEPTH, -math.MaxInt, math.MaxInt, false, preval)
		return
	} else if flag == 3 {
		return minimax_quiescence(game, DEPTH, -math.MaxInt, math.MaxInt, false, preval)
	} else if flag == 2 {
		return minimax_alpha_beta(game, DEPTH, -math.MaxInt, math.MaxInt, false, preval)
	} else {
		return minimax_plain(game, DEPTH, false, preval)
	}
}

func init_explored_depth() {
	for i := 0; i < len(explored_depth); i++ {
		explored_depth[i] = 0
	}
}

func init_hash_count() {
	for i := 0; i < 3; i++ {
		hash_count_list[i] = 0
	}
}

func deepening_counts(total_hash *int, total_explored *int, total_hash_list *[3]int, total_explored_list *[50]int) {
	*total_hash += hash_count
	*total_explored += explored
	total_hash_list[0] += hash_count_list[0]
	total_hash_list[1] += hash_count_list[1]
	total_hash_list[2] += hash_count_list[2]
	for i, v := range total_explored_list {
		v += explored_depth[i]
	}
	hash_count = 0
	explored = 0
	init_explored_depth()
	init_hash_count()
}

func get_pos_val(piece chess.PieceType, x int8, y int8, max bool) int {
	types := chess.PieceTypes()

	if max {
		switch piece {
		case types[0]:
			return pos_k[y][x]
		case types[1]:
			return pos_q[y][x]
		case types[2]:
			return pos_r[y][x]
		case types[3]:
			return pos_b[y][x]
		case types[4]:
			return pos_n[y][x]
		case types[5]:
			return pos_p[y][x]
		}
	} else {
		switch piece {
		case types[0]:
			return pos_k[7-y][x]
		case types[1]:
			return pos_q[7-y][x]
		case types[2]:
			return pos_r[7-y][x]
		case types[3]:
			return pos_b[7-y][x]
		case types[4]:
			return pos_n[7-y][x]
		case types[5]:
			return pos_p[7-y][x]
		}
	}

	// throw an error
	return 0
}

// ------ constants -------

func PieceValue(p chess.PieceType) int {
	types := chess.PieceTypes()
	switch p {
	case types[0]:
		return 60000
	case types[1]:
		return 929
	case types[2]:
		return 479
	case types[3]:
		return 320
	case types[4]:
		return 280
	case types[5]:
		return 100
	}
	return -1
}

var pos_p = [8][8]int{
	{100, 100, 100, 100, 105, 100, 100, 100},
	{78, 83, 86, 73, 102, 82, 85, 90},
	{7, 29, 21, 44, 40, 31, 44, 7},
	{-17, 16, -2, 15, 14, 0, 15, -13},
	{-26, 3, 10, 9, 6, 1, 0, -23},
	{-22, 9, 5, -11, -10, -2, 3, -19},
	{-31, 8, -7, -37, -36, -14, 3, -31},
	{0, 0, 0, 0, 0, 0, 0, 0},
}
var pos_n = [8][8]int{
	{-66, -53, -75, -75, -10, -55, -58, -70},
	{-3, -6, 100, -36, 4, 62, -4, -14},
	{10, 67, 1, 74, 73, 27, 62, -2},
	{24, 24, 45, 37, 33, 41, 25, 17},
	{-1, 5, 31, 21, 22, 35, 2, 0},
	{-18, 10, 13, 22, 18, 15, 11, -14},
	{-23, -15, 2, 0, 2, 0, -23, -20},
	{-74, -23, -26, -24, -19, -35, -22, -69},
}
var pos_b = [8][8]int{
	{-59, -78, -82, -76, -23, -107, -37, -50},
	{-11, 20, 35, -42, -39, 31, 2, -22},
	{-9, 39, -32, 41, 52, -10, 28, -14},
	{25, 17, 20, 34, 26, 25, 15, 10},
	{13, 10, 17, 23, 17, 16, 0, 7},
	{14, 25, 24, 15, 8, 25, 20, 15},
	{19, 20, 11, 6, 7, 6, 20, 16},
	{-7, 2, -15, -12, -14, -15, -10, -10},
}
var pos_r = [8][8]int{
	{35, 29, 33, 4, 37, 33, 56, 50},
	{55, 29, 56, 67, 55, 62, 34, 60},
	{19, 35, 28, 33, 45, 27, 25, 15},
	{0, 5, 16, 13, 18, -4, -9, -6},
	{-28, -35, -16, -21, -13, -29, -46, -30},
	{-42, -28, -42, -25, -25, -35, -26, -46},
	{-53, -38, -31, -26, -29, -43, -44, -53},
	{-30, -24, -18, 5, -2, -18, -31, -32},
}
var pos_q = [8][8]int{
	{6, 1, -8, -104, 69, 24, 88, 26},
	{14, 32, 60, -10, 20, 76, 57, 24},
	{-2, 43, 32, 60, 72, 63, 43, 2},
	{1, -16, 22, 17, 25, 20, -13, -6},
	{-14, -15, -2, -5, -1, -10, -20, -22},
	{-30, -6, -13, -11, -16, -11, -16, -27},
	{-36, -18, 0, -19, -15, -15, -21, -38},
	{-39, -30, -31, -13, -31, -36, -34, -42},
}
var pos_k = [8][8]int{
	{4, 54, 47, -99, -99, 60, 83, -62},
	{-32, 10, 55, 56, 56, 55, 10, 3},
	{-62, 12, -57, 44, -67, 28, 37, -31},
	{-55, 50, 11, -4, -19, 13, 0, -49},
	{-55, -43, -52, -28, -51, -47, -8, -50},
	{-47, -42, -43, -79, -64, -32, -29, -32},
	{-4, 3, -14, -50, -57, -18, 13, 4},
	{17, 30, -3, -14, 6, -1, 40, 18},
}
