package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
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

*/
const flag int = 3

const DO_MOVE_ORDERING bool = true

var DO_ITERATIVE_DEEPENING bool = true
var TIME_TO_THINK int = 2
var DEPTH int = 4 // default value without iterative deepening
var MAX_MOVES = int(math.Inf(1))
var MAX_QUIESCENCE = -10

var explored int = 0
var hash_count int = 0
var hash_count_list = [3]int{0, 0, 0}
var explored_depth [20]int
var position_eval = 0
var move_count = 0

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

// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
// rn1r2k1/ppp3pp/8/2b2b2/4P2q/2P1P3/PP1KQ1BP/RN4NR w - - 0 3
var start_pos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func main() {
	eng, err := uci.New("stockfish")
	if err != nil {
		panic(err)
	}
	defer eng.Close()
	// initialize uci with new game
	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		panic(err)
	}

	setup()
	fen, _ := chess.FEN(start_pos)
	game := chess.NewGame(fen)
	for game.Outcome() == chess.NoOutcome && move_count < MAX_MOVES {
		fmt.Println("Turn is now", game.Position().Turn())
		var move *chess.Move
		if game.Position().Turn() == chess.Black {
			start := time.Now()

			move = evaluate(game)
			game.Move(move)

			fmt.Println(game.Position())
			fmt.Println(game)
			end := time.Now()
			fmt.Println("Time elapsed", end.Sub(start))
		} else {
			// moves := game.ValidMoves()
			// move = moves[0]
			cmdPos := uci.CmdPosition{Position: game.Position()}
			cmdGo := uci.CmdGo{MoveTime: time.Second / 100}
			if err := eng.Run(cmdPos, cmdGo); err != nil {
				panic(err)
			}
			move := eng.SearchResults().BestMove
			if err := game.Move(move); err != nil {
				panic(err)
			}
		}

		pre := game.Clone()
		game.Move(move)
		position_eval = evaluate_position(pre, game, position_eval, move)
		fmt.Println(game.Position().Board().Draw())
		fmt.Println(game.Position())
		fmt.Println()
		move_count++
	}
	// print outcome and game PGN
	fmt.Printf("Game completed. %s by %s.\n", game.Outcome(), game.Method())
	fmt.Println(game)
}

func setup() {
	fmt.Println("Start game...")
	init_explored_depth()
	init_hash_count()
	generateZobristConstants()
}

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

func zobrist(game *chess.Game, max bool) uint64 {
	var bits uint64 = 0
	pos := game.Position().Board().SquareMap()
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
			return 2, 0, nil, p.moves
		} else {
			fmt.Println("HASH CONFLICT", hash, p.hash, p)
		}
	}
	return 3, 0, nil, nil
}

func quiescence_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int, move_gen []*chess.Move) (best *chess.Move, eval int) {
	moves := get_quiescence_moves(game, move_gen)

	if len(moves) == 0 {
		write_hash(zobrist(game, max), depth, "EDGE", preval, nil, nil, game.Position())
		return nil, preval
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
			}
			if tempeval > alpha {
				alpha = tempeval
			}
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
			}
			if tempeval < beta {
				beta = tempeval
			}
			if alpha >= beta {
				break
			}

		}
	}
	return
}

func minimax_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int) (best *chess.Move, eval int) {

	if depth < MAX_QUIESCENCE {
		write_hash(zobrist(game, max), depth, "EDGE", preval, nil, nil, game.Position())
		return nil, preval
	}

	var moves []*chess.Move
	flag, hashscore, hashbest, hashmoves := read_hash(zobrist(game, max), depth, alpha, beta)

	if flag == 1 {
		return hashbest, hashscore
	} else if flag == 2 {
		moves = hashmoves
	} else {
		moves = game.ValidMoves()
	}

	explored++
	explored_depth[DEPTH-depth]++

	// makes sure we don't run quiescence move pruning on empty
	if len(moves) == 0 {
		write_hash(zobrist(game, max), depth, "EDGE", preval, nil, nil, game.Position())
		return nil, preval
	}

	if depth <= 0 {
		return quiescence_hashing(game, depth, alpha, beta, max, preval, moves)
	}

	// this avoids ordering moves if quiescence
	if flag != 2 {
		moves = move_order(game, moves)
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
			}
			if tempeval > alpha {
				alpha = tempeval
			}
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
				if depth == DEPTH {
					fmt.Println()
					fmt.Println(move, tempeval, beta)
				}
			}
			if tempeval < beta {
				beta = tempeval
			}
			if alpha >= beta {
				break
			}
			if depth == DEPTH {
				fmt.Print("x")
			}
		}
	}
	if depth == DEPTH {
		fmt.Print("x")
	}

	if max {
		write_hash(zobrist(game, max), depth, "ALPHA", alpha, best, moves, game.Position())
	} else {
		write_hash(zobrist(game, max), depth, "BETA", beta, best, moves, game.Position())
	}
	return
}

func evaluate_position(pre *chess.Game, post *chess.Game, preval int, move *chess.Move) (eval int) {

	eval = preval
	if move == nil { // first round evaluation
		return position_eval
	}

	var flip int = 1
	max := pre.Position().Turn() == chess.White
	if !max {
		flip = -1
	}

	if post.Outcome() == chess.WhiteWon {
		return 1000000
	}
	if post.Outcome() == chess.BlackWon {
		return -1000000
	}
	if post.Outcome() == chess.Draw {
		return 0
	}

	if move.HasTag(chess.Check) {
		eval += flip * 50
	}

	if move.HasTag(chess.Capture) {
		// fmt.Println(PieceValue(game.Position().Board().Piece(move.S2()).Type()))
		eval += flip * PieceValue(pre.Position().Board().Piece(move.S2()).Type())
	}

	move_type := pre.Position().Board().Piece(move.S1()).Type()
	from := get_pos_val(move_type, int8(move.S1().File()), int8(move.S1().Rank()), max)
	to := get_pos_val(move_type, int8(move.S2().File()), int8(move.S2().Rank()), max)
	eval += flip * (to - from)

	return eval
}

func get_quiescence_moves(game *chess.Game, moves []*chess.Move) []*chess.Move {

	funcVar := func(move *chess.Move) bool {
		if move.HasTag(chess.Capture) {
			return true
		}
		if move.HasTag(chess.Check) {
			post := game.Clone()
			post.Move(move)
			if post.Outcome() == chess.WhiteWon || post.Outcome() == chess.BlackWon {
				return true
			}
		}
		if move.Promo().String() != "" {
			return true
		}
		return false
	}

	result := make([]*chess.Move, 0, len(moves))
	for _, move := range moves {
		if funcVar(move) {
			result = append(result, move)
		}
	}

	return result
}

func move_order(game *chess.Game, moves []*chess.Move) []*chess.Move {
	// const len int = len(moves)
	evaluated := make(map[*chess.Move]int)
	for _, move := range moves {
		evaluated[move] = evaluate_move(game, move)
	}

	keys := make([]*chess.Move, 0, len(evaluated))
	for key := range evaluated {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return evaluated[keys[i]] > evaluated[keys[j]] })

	return keys
}

func evaluate_move(game *chess.Game, move *chess.Move) (eval int) {
	if move.Promo().String() != "" {
		return 2000
	}

	eval = 0
	max := game.Position().Turn() == chess.White
	if move.HasTag(chess.Check) {
		eval += 1000
	}

	move_type := game.Position().Board().Piece(move.S1()).Type()

	if move.HasTag(chess.Capture) {
		eval += PieceValue(game.Position().Board().Piece(move.S2()).Type()) - PieceValue(move_type)
	}

	from := get_pos_val(move_type, int8(move.S1().File()), int8(move.S1().Rank()), max)
	to := get_pos_val(move_type, int8(move.S2().File()), int8(move.S2().Rank()), max)
	eval += to - from

	return
}

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

func evaluate(game *chess.Game) (output *chess.Move) {
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
	DEPTH = 1
	for time.Now().Sub(delay) < 0 {
		print_iter_1(delay)
		output, _ = minimax_factory(game, 0)
		DEPTH++
		print_iter_2(output, 0)
		hash_count = 0
		explored = 0
		init_explored_depth()
		init_hash_count()
		
	}
	return
}

// --------- old code ------------

func quiescence(game *chess.Game, depth int, alpha int, beta int, max bool, preval int, move_gen []*chess.Move) (best *chess.Move, eval int) {
	moves := get_quiescence_moves(game, move_gen)

	if len(moves) == 0 {
		return nil, preval
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_quiescence(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
			}
			if tempeval > alpha {
				alpha = tempeval
			}
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_quiescence(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
			}
			if tempeval < beta {
				beta = tempeval
			}
			if alpha >= beta {
				break
			}

		}
	}
	return
}

func minimax_quiescence(game *chess.Game, depth int, alpha int, beta int, max bool, preval int) (best *chess.Move, eval int) {
	// fmt.Println(depth)
	explored++
	explored_depth[DEPTH-depth]++

	move_gen := game.ValidMoves()
	if depth < MAX_QUIESCENCE {
		return nil, preval
	}

	if depth <= 0 {
		return quiescence(game, depth, alpha, beta, max, preval, move_gen)
	}

	moves := move_order(game, move_gen)

	if len(moves) == 0 {
		return nil, preval
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_quiescence(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
			}
			if tempeval > alpha {
				alpha = tempeval
			}
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_quiescence(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
			}
			if tempeval < beta {
				beta = tempeval
			}
			if alpha >= beta {
				break
			}
			if depth == DEPTH {
				fmt.Println(move, tempeval, beta)
			}

		}
	}

	return
}

func minimax_alpha_beta(game *chess.Game, depth int, alpha int, beta int, max bool, preval int) (best *chess.Move, eval int) {
	// fmt.Println(depth)
	explored++
	explored_depth[DEPTH-depth]++

	if depth == 0 {
		return nil, preval
	}

	move_gen := game.ValidMoves()
	moves := move_gen
	if DO_MOVE_ORDERING {
		moves = move_order(game, move_gen)
	}

	if len(moves) == 0 {
		return nil, preval
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_alpha_beta(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
			}
			if tempeval > alpha {
				alpha = tempeval
			}
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_alpha_beta(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
				if depth == DEPTH {
					fmt.Println(move, tempeval, beta)
				}
			}
			if tempeval < beta {
				beta = tempeval
			}
			if alpha >= beta {
				break
			}

		}
	}
	return
}

func minimax_plain(game *chess.Game, depth int, max bool, preval int) (best *chess.Move, eval int) {
	// fmt.Println(depth)
	explored++
	explored_depth[DEPTH-depth]++

	if depth == 0 {
		return nil, preval
	}

	move_gen := game.ValidMoves()
	moves := move_gen

	if len(moves) == 0 {
		return nil, preval
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_plain(post, depth-1, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
			}
		}
	} else {
		eval = math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval := minimax_plain(post, depth-1, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
				if depth == DEPTH {
					fmt.Println(move, tempeval)
				}
			}
		}
	}

	return
}

func capture_only(game *chess.Game) *chess.Move {
	moves := game.ValidMoves()
	for _, value := range moves {
		if value.HasTag(chess.Capture) {
			return value
		}
	}
	// fmt.Println(moves)
	move := moves[rand.Intn(len(moves))]
	return move
}

// ----- print statements to clean up code ----

func print_iter_1(delay time.Time) {
	fmt.Println("new interation with depth: ", DEPTH, "time left:", delay.Sub(time.Now()))
}

func print_iter_2(output *chess.Move, eval int) {
	fmt.Println("done", output, eval, "\n")
	fmt.Println(hash_count)
	fmt.Println(hash_count_list)
	fmt.Println(explored)
	fmt.Println(explored_depth)
}

// ------ now entering the doldrums -----

func minimax_factory(game *chess.Game, preval int) (best *chess.Move, eval int) {
	if flag == 4 {
		return minimax_hashing(game, DEPTH, -math.MaxInt, math.MaxInt, false, preval)
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
