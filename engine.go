package main

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/notnil/chess"
)

/*

Todo:

Fix evaluation functions, static evals, etc.
Investigate remaining insanity. (checks off via quiescence? hashing problem likely)
Opening Books
Parralelization
Null Window Search (AKA Negascout/PVS)

*/

func minimax_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int) (best *chess.Move, eval int, history [mem_size]*chess.Move) {
	index_depth := DEPTH - depth
	explored++
	explored_depth[index_depth]++

	if check_time_up() {
		_, hashscore, hashbest, _, _ := read_hash(zobrist(game.Position().Board(), max), int(math.Inf(-1)), int(math.Inf(-1)), int(math.Inf(1)))
		history[index_depth] = hashbest
		return hashbest, hashscore, history
	}

	if depth <= MAX_QUIESCENCE || index_depth >= MAX_DEPTH {
		return end_at_edge(game, depth, max, preval)
	}

	flag, hashscore, hashbest, hashmoves, depthfound := read_hash(zobrist(game.Position().Board(), max), depth, alpha, beta)

	if flag == DeeperResult {
		history[index_depth] = hashbest
		return hashbest, hashscore, history
	}

	var moves []*chess.Move
	if depth <= 0 { 
		if flag == QuiescenceDeeperResult {
			history[index_depth] = hashbest
			return hashbest, hashscore, history
		} else if flag == QuiescenceSavedMoves {
			moves = hashmoves
		} else { 
			if flag == MinimaxSavedMoves {
				moves = hashmoves
			} else {
				moves = game.ValidMoves()
			}
			moves = get_quiescence_moves(game, moves)
		}

		if len(moves) == 0 { // if quiet
			return end_at_edge(game, depth, max, preval)
		} else { // not quiet
			// fmt.Println(alpha, beta, preval, max)
			if max {
				if preval >= alpha {
					return end_at_edge(game, depth, max, preval)
				}
				return quiescence_hashing(game, depth, alpha, preval, max, preval, moves)

			} else {
				if preval <= beta {
					return end_at_edge(game, depth, max, preval)
				}
				return quiescence_hashing(game, depth, preval, beta, max, preval, moves)
			}
		}
	}



	if flag == MinimaxSavedMoves {
		moves = hashmoves
	} else {
		moves = game.ValidMoves()
	}

	if len(moves) == 0 {
		return end_at_edge(game, depth, max, preval)
	}

	if flag != 2 {
		moves = move_order(game, moves)
	}
	
	root := index_depth == 0

	if root && VERBOSE_PRINT {
		fmt.Println("\nDEPTH:", depth, preval)
		// fmt.Println("MOVE ORDER:\n", moves)
		fmt.Println("HASH RETURN:\n", depthfound, hashscore, "\n")
		// hashbest, hashmoves, flag
		// fmt.Println(hash_map[zobrist(game.Position().Board(), max)])
	}

	return minimax_hashing_core(game, depth, alpha, beta, max, preval, moves)
}

func minimax_hashing_core(game *chess.Game, depth int, alpha int, beta int, max bool, preval int, moves []*chess.Move) (best *chess.Move, eval int, history [mem_size]*chess.Move) {
	root := depth == DEPTH
	index_depth := DEPTH - depth
	move_sorting := make(map[*chess.Move]int)

	if max {
		eval = -1 * math.MaxInt
		for i, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval != tempeval && check_time_up() {
				continue
			}
			move_sorting[move] = tempeval
			if tempeval > eval {
				eval = tempeval
				best = move
				temphistory[index_depth] = move
				history = temphistory
				if root {
					print_root_move_1(post, move, tempeval, alpha, history)
				}
			}
			if tempeval > alpha {
				alpha = tempeval
			}
			if tempeval >= 1000000 {
				// fmt.Println("CHECKMATE FOR WHITE")
				// fmt.Println(post, move, eval, tempeval, alpha, history)
				for _, move := range moves[i:] {
					move_sorting[move] = -10000
				}
				break
			}
			if alpha >= beta {
				for _, move := range moves[i:] {
					move_sorting[move] = -10000
				}
				break
			}
		}
	} else {
		eval = math.MaxInt
		for i, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval != tempeval {
				if check_time_up() {
					continue
				} else {
					panic("NOT A NUMBER")
				}
			} 
			move_sorting[move] = tempeval
			if tempeval < eval {
				eval = tempeval
				best = move
				temphistory[index_depth] = move
				history = temphistory
				if root {
					print_root_move_1(post, move, tempeval, beta, history)
				}
				
			} else {
				if root {
					fmt.Print("x")
				}
			}
			if tempeval < beta {
				beta = tempeval
			}
			if tempeval <= -1000000 {
				// fmt.Println("CHECKMATE FOR BLACK")
				// fmt.Println(post, move, eval, tempeval, beta, history)
				for _, move := range moves[i:] {
					move_sorting[move] = 10000
				}
				break
			}
			if alpha >= beta {
				for _, move := range moves[i:] {
					move_sorting[move] = 10000
				}
				break
			}
		}
	}

	m := make([]*chess.Move, 0, len(move_sorting))
	for move := range move_sorting {
		m = append(m, move)
	}

	// if len(m) != len(moves) {
	// 	fmt.Println(m, moves)
	// 	panic("WOAH")
	// }

	if max {
		sort.Slice(m, func(i, j int) bool { return move_sorting[m[i]] > move_sorting[m[j]] })
		write_hash(zobrist(game.Position().Board(), max), depth, AlphaFlag, alpha, best, m, game.Position())
	} else {
		sort.Slice(m, func(i, j int) bool { return move_sorting[m[i]] < move_sorting[m[j]] })
		write_hash(zobrist(game.Position().Board(), max), depth, BetaFlag, beta, best, m, game.Position())
	}
	if root {
		fmt.Print("\n")
		if check_time_up() {
			fmt.Println("\n -- Returned early because time was up... -- ")
		}
	}
	return
}

func quiescence_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int, moves []*chess.Move) (best *chess.Move, eval int, history [mem_size]*chess.Move) {
	if max {
		eval = int(math.Inf(-1))
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
				temphistory[DEPTH-depth] = move
				history = temphistory
			}
			if tempeval > alpha {
				alpha = tempeval
			}
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = int(math.Inf(1))
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
				temphistory[DEPTH-depth] = move
				history = temphistory
			}
			if tempeval < beta {
				beta = tempeval
			}
			if alpha >= beta {
				break
			}
		}
	}
	
	if max {
		write_hash(zobrist(game.Position().Board(), max), depth, AlphaQFlag, alpha, best, moves, game.Position())
	} else {
		write_hash(zobrist(game.Position().Board(), max), depth, BetaQFlag, beta, best, moves, game.Position())
	}
	if best == nil {
		return end_at_edge(game, depth, max, preval)
	}
	return
}

func check_time_up() bool {
	if !DO_ITERATIVE_DEEPENING {
		return false
	}
	return delay.Sub(time.Now()) < 0
}

func end_at_edge(game *chess.Game, depth int, max bool, preval int) (best *chess.Move, eval int, history [mem_size]*chess.Move) {
	write_hash(zobrist(game.Position().Board(), max), depth, EdgeFlag, preval, nil, nil, game.Position())
	return nil, preval, history // history is blank
}

// -------------------------

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
		eval += flip * 20
	}

	if move.HasTag(chess.Capture) {
		// fmt.Println(PieceValue(game.Position().Board().Piece(move.S2()).Type()))
		eval += flip * PieceValue(pre.Position().Board().Piece(move.S2()).Type())
	}

	move_type := pre.Position().Board().Piece(move.S1()).Type()
	from := get_pos_val(move_type, int8(move.S1().File()), int8(move.S1().Rank()), max)
	to := get_pos_val(move_type, int8(move.S2().File()), int8(move.S2().Rank()), max)
	eval += flip * (to - from) / 10

	return eval
}

func get_quiescence_moves(game *chess.Game, moves []*chess.Move) []*chess.Move {
	funcVar := func(move *chess.Move) bool {
		if move.HasTag(chess.Capture) {
			return true
		}
		if move.HasTag(chess.Check) {
			// post := game.Clone()
			// post.Move(move)
			// if post.Outcome() == chess.WhiteWon || post.Outcome() == chess.BlackWon {
			return true
			// }
		}
		if move.Promo() != chess.PieceType(0) {
			return true
		}
		return false
	}

	evaluated := make(map[*chess.Move]int)
	for _, move := range moves {
		if funcVar(move) {
			evaluated[move] = evaluate_quiescence_move(game, move)
		}
	}

	result := make([]*chess.Move, 0, len(evaluated))
	for key := range evaluated {
		result = append(result, key)
	}
	sort.Slice(result, func(i, j int) bool { return evaluated[result[i]] > evaluated[result[j]] })

	return result
}

func move_order(game *chess.Game, moves []*chess.Move) []*chess.Move {
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
	if move.Promo() != chess.PieceType(0) {
		return 2000
	}
	
	eval = 0
	max := game.Position().Turn() == chess.White

	move_type := game.Position().Board().Piece(move.S1()).Type()

	if move.HasTag(chess.Capture) {
		eval += PieceValue(game.Position().Board().Piece(move.S2()).Type()) - PieceValue(move_type)
	} else if move.HasTag(chess.Check) {
		eval += 10
	}

	from := get_pos_val(move_type, int8(move.S1().File()), int8(move.S1().Rank()), max)
	to := get_pos_val(move_type, int8(move.S2().File()), int8(move.S2().Rank()), max)
	eval += (to - from) / 10

	return
}

func evaluate_quiescence_move(game *chess.Game, move *chess.Move) int {
	move_type := game.Position().Board().Piece(move.S1()).Type()
	return PieceValue(game.Position().Board().Piece(move.S2()).Type()) - PieceValue(move_type)
}
