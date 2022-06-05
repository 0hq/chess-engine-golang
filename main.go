package main

import (
	"fmt"
	"math"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
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
Opening Book?
Parralel Search

*/
const flag int = 4

const DO_MOVE_ORDERING bool = true
const DO_ITERATIVE_DEEPENING bool = true
const DO_STRICT_TIMING bool = false
const MAX_ITERATIVE_DEPTH int = 200
const TIME_TO_THINK int = 10
const MAX_MOVES = 200
const MAX_QUIESCENCE = -1000
var VERBOSE_PRINT = true

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
// r1bqkb1r/ppp1ppp1/1Pnp4/4P3/2BP3p/2N2N1P/PP3PP1/R1BQK2R w KQkq - 0 10 blunder 
// "r2qkb2/1pp1p1p1/1pn2p2/4pb2/2BP3r/2N2N1P/PP1K1PP1/R2Q3R w q - 0 15" // f3h4 d8d4 c4d3 d4d3
// quiescence testing good 4k3/5q2/1p6/1Pp5/2Pp4/3Pp3/4P3/1K3R2 w - - 0 1
// quiescence testing bad 4k3/8/1p6/1Pp5/2Pp4/3Pp3/4P3/1K3Q2 w - - 0 1
// 3rq1k1/4br1p/2ppb1p1/p5Pn/N3P2P/4Q3/1PP3BB/1K1RR3 w - - 1 22
// rnbqkbnr/ppp2ppp/8/3pp3/8/6PP/PPPPPP2/RNBQKBNR w KQkq d6 0 3
var start_pos = "r3kb1r/1b2pppp/p3q3/3N4/8/5B2/PPP2PPP/R2Q1RK1 b kq - 6 12"

func main() {
	run_tests()
	game := setup()

	// initialize stockfish
	eng, err := uci.New("stockfish")
	if err != nil {
		panic(err)
	}
	defer eng.Close()
	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		panic(err)
	}

	for game.Outcome() == chess.NoOutcome && move_count < MAX_MOVES {
		var move *chess.Move
		color := game.Position().Turn()
		start := time.Now()

		fmt.Println("\n\n", move_count, "---- New Turn ----", color)

		if color == engine_color {
			move = engine(game, engine_color == chess.White)
		} else {
			// move = engine(game)
			// move = random_move_engine(game)
			move = stockfish(game, eng)
		}

		if move == nil {
			panic("NO MOVE")
		}

		pre := game.Clone()
		game.Move(move)

		print_turn_complete(game, move, start)
		update_evaluation(game, pre, move)
		move_count++
	}
	print_game_over(game)
}

func run_tests() {
	fmt.Println("\nRunning tests...")
	VERBOSE_PRINT = false
	init_explored_depth()
	init_hash_count()
	generateZobristConstants()
	fen, _ := chess.FEN("3qr2k/pbpp2pp/1p5N/3Q2b1/2P1P3/P7/1PP2PPP/R4RK1 w - - 0 1")
	game := chess.NewGame(fen)
	move := engine(game, true) 
	// fmt.Println(move)
	if move.String() != "d5g8" {
		panic("TEST FAILED")
	}
	VERBOSE_PRINT = true
	fmt.Println("Tests passed...\n")
}

func setup() *chess.Game {
	fmt.Println("\n\nStart game...")
	init_explored_depth()
	init_hash_count()
	generateZobristConstants()
	fen, _ := chess.FEN(start_pos)
	return chess.NewGame(fen)
}

func engine(game *chess.Game, max bool) (output *chess.Move) {
	explored = 0
	init_explored_depth()
	if DO_ITERATIVE_DEEPENING {
		output = iterative_deepening(game, TIME_TO_THINK, max)
	} else {
		var history [mem_size]string
		output, _, history = minimax_factory(game, 0, max)
		fmt.Println(history)
		print_iter_2()
	}
	return
}

func iterative_deepening(game *chess.Game, time_control int, max bool) (output *chess.Move) {
	delay = time.Now()
	delay = delay.Add(time.Second * time.Duration(time_control))
	DEPTH = 1 // starting depth

	// var total_hash, total_explored int
	// var total_hash_list [3]int
	// var total_explored_list [mem_size]int
	var eval int

	for time.Now().Sub(delay) < 0 {
		print_iter_1(delay)
		var history [mem_size]string
		output, eval, history = minimax_factory(game, 0, max)
		fmt.Println("\n", output)
		fmt.Println("line:", history)
		fmt.Println("evaluation:", eval)
		print_iter_2()
		// deepening_counts(&total_hash, &total_explored, &total_hash_list, &total_explored_list)
		DEPTH++
		if eval >= 10000 || eval <= -10000 || DEPTH > MAX_ITERATIVE_DEPTH {
			break
		}
	}

	// fmt.Println("\n\nTotal nodes explored", total_explored)
	// fmt.Println("# nodes at depth", explored_depth)
	// fmt.Println("Total hashes used", total_hash)
	// fmt.Println("Hashes written", eval)
	// fmt.Println("Hash types (edge, alpha, beta)", total_hash_list)
	return
}

// this should be static, not relative
func update_evaluation(game *chess.Game, pre *chess.Game, move *chess.Move) {
	position_eval = evaluate_position(pre, game, position_eval, move)
}

// ----- print statements to clean up code ----

func print_root_move_1(game *chess.Game, move *chess.Move, tempeval int, cap int, history [mem_size]string) {
	if !VERBOSE_PRINT {
		return
	}
	fmt.Println("\nNew best root move:", move)
	fmt.Println("Evaluation:", tempeval, "Prev eval (forced beta/alpha):", cap)
	fmt.Println("Move path:", history)
	// fmt.Println(game.Position().Board().Draw())
}

func print_iter_1(delay time.Time) {
	if !VERBOSE_PRINT {
		return
	}
	fmt.Println("\n -- Searching deeper --")
	fmt.Println("Depth:", DEPTH)
	fmt.Println("Time left:", delay.Sub(time.Now()), "\n")
}

func print_iter_2() {
	if !VERBOSE_PRINT {
		return
	}
	fmt.Println("\nTotal nodes explored", explored)
	fmt.Println("# nodes at depth", explored_depth)
	fmt.Println("Total hashes used", hash_count)
	fmt.Println("Hashes written", hash_write_count)
	fmt.Println("Hash types (edge, alpha, beta)", hash_count_list)
}

func print_turn_complete(game *chess.Game, move *chess.Move, start time.Time) {
	if !VERBOSE_PRINT {
		return
	}
	fmt.Println("\nMove chosen:", move)
	end := time.Now()
	fmt.Println(game.Position().Board().Draw())
	fmt.Println("Time elapsed", end.Sub(start))
	fmt.Println(`[SetUp "1"]`)
	fmt.Print(`[FEN "`,start_pos,`"]`,"\n")
	fmt.Println(game)
	// fmt.Println(game.Position())
	// fmt.Println(game)
}

func print_game_over(game *chess.Game) {
	if !VERBOSE_PRINT {
		return
	}
	fmt.Printf("\n\n ----- Game completed. %s by %s. ------\n\n", game.Outcome(), game.Method())
	fmt.Println(`[SetUp "1"]`)
	fmt.Print(`[FEN "`,start_pos,`"]`,"\n")
	fmt.Println(game)
	
}

// ------ now entering the doldrums -----

func minimax_factory(game *chess.Game, preval int, max bool) (best *chess.Move, eval int, history [mem_size]string) {
	if flag == 4 {
		best, eval, history, _ = minimax_hashing(game, DEPTH, -math.MaxInt, math.MaxInt, max, preval)
	} else if flag == 3 {
		best, eval = minimax_quiescence(game, DEPTH, -math.MaxInt, math.MaxInt, max, preval)
	} else if flag == 2 {
		best, eval =  minimax_alpha_beta(game, DEPTH, -math.MaxInt, math.MaxInt, max, preval)
	} else if flag == 1 {
		best, eval =  minimax_plain(game, DEPTH, max, preval)
	}
	return
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

func deepening_counts(total_hash *int, total_explored *int, total_hash_list *[3]int, total_explored_list *[mem_size]int) {
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
	return 0

	types := chess.PieceTypes()


	if max {
		switch piece {
		case types[0]:
			return pos_k[y][x] / 10 // king pos is disabled
		case types[1]:
			return pos_q[y][x]
		case types[2]:
			return pos_r[y][x] / 2
		case types[3]:
			return pos_b[y][x]
		case types[4]:
			return pos_n[y][x]
		case types[5]:
			return pos_p[y][x] * 2
		}
	} else {
		switch piece {
		case types[0]:
			return pos_k[7-y][x] / 10 // king pos is disabled
		case types[1]:
			return pos_q[7-y][x]
		case types[2]:
			return pos_r[7-y][x] / 2
		case types[3]:
			return pos_b[7-y][x]
		case types[4]:
			return pos_n[7-y][x]
		case types[5]:
			return pos_p[7-y][x] * 2
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
	{4, 50, 60, -99, -99, 70, 90, -62},
	{-32, 10, 55, 56, 56, 55, 10, 3},
	{-62, 12, -57, 44, -67, 28, 37, -31},
	{-55, 50, 11, -4, -19, 13, 0, -49},
	{-55, -43, -52, -28, -51, -47, -8, -50},
	{-47, -42, -43, -79, -64, -32, -29, -32},
	{-4, 3, -14, -50, -57, -18, 13, 4},
	{17, 30, -3, -14, 6, -1, 40, 18},
}
