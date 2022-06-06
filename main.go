package main

import (
	"fmt"
	"math"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
)

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
			// move = engine(game, engine_color == chess.Black)
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

func iterative_deepening_mtdf(game *chess.Game, time_control int, max bool) (output *chess.Move) {

	DEPTH = 1 // starting depth
	delay = time.Now().Add(time.Second * time.Duration(time_control))
	var eval int = 0
	var history [mem_size]string

	for time.Now().Sub(delay) < 0 {
		fmt.Println("\n\nnew depth", DEPTH)
		
		print_iter_1(delay)

		output, eval, history = mtdf_algo(game, DEPTH, max, eval)
		
		print_iter_11(output, eval, history)
		print_iter_2()
		
		DEPTH++
		if eval >= 10000 || eval <= -10000 || DEPTH > MAX_ITERATIVE_DEPTH {
			break
		}
	}

	return
}

func iterative_deepening(game *chess.Game, time_control int, max bool) (output *chess.Move) {

	DEPTH = 1 // starting depth
	delay = time.Now().Add(time.Second * time.Duration(time_control))
	var eval int
	var history [mem_size]string

	for time.Now().Sub(delay) < 0 {
		
		print_iter_1(delay)

		output, eval, history = minimax_factory(game, 0, max)
		
		print_iter_11(output, eval, history)
		print_iter_2()
		
		DEPTH++
		if eval >= 10000 || eval <= -10000 || DEPTH > MAX_ITERATIVE_DEPTH {
			break
		}
	}

	return
}

func run_tests() {
	fmt.Println("\nRunning tests...")
	stored := VERBOSE_FLAG
	VERBOSE_FLAG = 0
	init_explored_depth()
	init_hash_count()
	generateZobristConstants()
	fen, _ := chess.FEN("3qr2k/pbpp2pp/1p5N/3Q2b1/2P1P3/P7/1PP2PPP/R4RK1 w - - 0 1")
	game := chess.NewGame(fen)
	move := engine(game, true)
	fmt.Println(move)
	if move.String() != "d5g8" {
		panic("TEST FAILED")
	}
	VERBOSE_FLAG = stored
	fmt.Println("Tests passed...\n")
}

func setup() *chess.Game {
	fmt.Println("\n\nStart game...")
	opening_moves = true
	init_explored_depth()
	init_hash_count()
	generateZobristConstants()
	fen, _ := chess.FEN(start_pos)
	return chess.NewGame(fen)
}

func engine(game *chess.Game, max bool) (output *chess.Move) {

	if opening_moves {
		move := get_opening(game, 0)
		fmt.Println(move)
		if move == nil {
			opening_moves = false
		} else {
			return move
		}
		// panic("te")
	}


	explored = 0
	init_explored_depth()
	if DO_MTDF {
		output = iterative_deepening_mtdf(game, TIME_TO_THINK, max)
	} else if DO_ITERATIVE_DEEPENING {
		output = iterative_deepening(game, TIME_TO_THINK, max)
	} else {
		var history [mem_size]string
		output, _, history = minimax_factory(game, 0, max)
		fmt.Println(history)
		print_iter_2()
	}
	return
}

func minimax_factory(game *chess.Game, preval int, max bool) (best *chess.Move, eval int, history [mem_size]string) {
	if flag == 4 {
		best, eval, history, _ = minimax_hashing(game, DEPTH, -math.MaxInt, math.MaxInt, max, preval)
	} else if flag == 3 {
		best, eval = minimax_quiescence(game, DEPTH, -math.MaxInt, math.MaxInt, max, preval)
	} else if flag == 2 {
		best, eval = minimax_alpha_beta(game, DEPTH, -math.MaxInt, math.MaxInt, max, preval)
	} else if flag == 1 {
		best, eval = minimax_plain(game, DEPTH, max, preval)
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
