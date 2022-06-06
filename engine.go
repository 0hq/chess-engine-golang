package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/notnil/chess"
)

/*

Todo:

Fix Hashing and Strict Timing
Parralelization
Null Window Search (AKA Negascout/PVS)

Improve evaluation function.
	Center Control, Isolation, King Safety, Mobility
Endgames?

To fix:
	En passant and castling in zobrist
	Draw detection
	King endgame table
	Strict timing is still broken

*/

var nullMove chess.Move = chess.Move{}


func minimax_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int) (best *chess.Move, eval int, history [mem_size]string, ignore bool) {
	index_depth := DEPTH - depth
	explored++
	explored_depth[index_depth]++

	if check_time_up() {
		// if it's top level send it back no matter what
		if index_depth == 1 {
			flag, hashscore, hashbest, _, depthfound := read_hash(zobrist(game.Position().Board(), max), int(math.Inf(-1)), int(math.Inf(-1)), int(math.Inf(1)))
			if flag == DeeperResult || flag == QuiescenceDeeperResult {
				if hashbest != nil {
					history[index_depth] = hashbest.String() + "- timeup top hashed" + strconv.Itoa(depthfound)
				} else {
					history[index_depth] = "timeup top hashed edge"
				}
				return hashbest, hashscore, history, false
			} else {
				history[index_depth] = "null time (top)"
				return nil, 0, history, true // don't use this calculation
			}
		}
		flag, hashscore, hashbest, _, depthfound := read_hash(zobrist(game.Position().Board(), max), depth, int(math.Inf(-1)), int(math.Inf(1)))
		if flag == DeeperResult || flag == QuiescenceDeeperResult {
			if hashbest != nil {
				history[index_depth] = hashbest.String() + "- timeup hashed" + strconv.Itoa(depthfound)
			} else {
				history[index_depth] = "timeup hashed edge"
			}
			return hashbest, hashscore, history, false
		} else {
			history[index_depth] = "null time"
			return nil, 0, history, true // don't use this calculation
		}
	}

	if depth <= MAX_QUIESCENCE || index_depth >= MAX_DEPTH {
		return end_at_edge(game, depth, max, preval)
	}

	flag, hashscore, hashbest, hashmoves, depthfound := read_hash(zobrist(game.Position().Board(), max), depth, alpha, beta)

	if flag == DeeperResult {
		if hashbest != nil {
			history[index_depth] = hashbest.String() + "-hdeeper"
		} else {
			history[index_depth] = "edge deeper"
		}
		return hashbest, hashscore, history, false
	}

	var moves []*chess.Move
	if depth <= 0 {
		if flag == QuiescenceDeeperResult {
			history[index_depth] = hashbest.String() + "-qdeeper"
			return hashbest, hashscore, history, false
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
			return quiescence_hashing(game, depth, alpha, beta, max, preval, moves)
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

	if root && VERBOSE_FLAG == 2 {
		fmt.Println("\nDEPTH:", depth, preval)
		// fmt.Println("MOVE ORDER:\n", moves)
		fmt.Println("HASH RETURN:\n", depthfound, hashscore, "\n")
		// hashbest, hashmoves, flag
		// fmt.Println(hash_map[zobrist(game.Position().Board(), max)])
	}

	return minimax_hashing_core(game, depth, alpha, beta, max, preval, moves)
}

func minimax_hashing_core(game *chess.Game, depth int, alpha int, beta int, max bool, preval int, moves []*chess.Move) (best *chess.Move, eval int, history [mem_size]string, ignore bool) {
	root := depth == DEPTH
	index_depth := DEPTH - depth
	move_sorting := make(map[*chess.Move]int)

	if max {
		eval = math.MinInt
		for _, move := range moves {

			// create a new game and simulate the move
			post := game.Clone()
			post.Move(move)

			// evaluate the position relatively (take current eval and take difference)
			state_eval := evaluate_position(game, post, preval, move)

			// search one depth further
			_, tempeval, temphistory, ignore := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)

			// save each move value
			move_sorting[move] = tempeval

			// this move wasn't evaluated due to time
			if ignore {
				continue
			}

			// save if better than previous move
			if tempeval > eval {
				eval = tempeval
				best = move
				temphistory[index_depth] = move.String()
				history = temphistory
				print_root_move_1(root, post, move, tempeval, beta, history)
			} else {
				print_root_move_2(root)
			}

			// set alpha is better than alpha
			if tempeval > alpha {
				alpha = tempeval
			}

			// checkmate for white
			if tempeval >= 1000000 {
				break
			}

			// there exists a preferrable path elsewhere that is always better for me
			// my opponent has an option in this branch that i can't avoid, not looking anymore
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = math.MaxInt
		for _, move := range moves {

			// create a new game and simulate the move
			post := game.Clone()
			post.Move(move)

			// evaluate the position relatively (take current eval and take difference)
			state_eval := evaluate_position(game, post, preval, move)

			// search one depth further
			_, tempeval, temphistory, ignore := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)

			// save each move value
			move_sorting[move] = tempeval

			// this move wasn't evaluated due to time
			if ignore {
				continue
			}

			// save if better than previous move
			if tempeval < eval {
				eval = tempeval
				best = move
				temphistory[index_depth] = move.String()
				history = temphistory
				print_root_move_1(root, post, move, tempeval, beta, history)
			} else {
				print_root_move_2(root)
			}

			// store the eval of my best move, so far
			if tempeval < beta {
				beta = tempeval
			}

			// checkmate for black
			if tempeval <= -1000000 {
				break
			}

			// there exists a preferrable path elsewhere that will always be better for me
			// my opponent has an option in this branch that i can't avoid, not looking anymore
			if alpha >= beta {
				break
			}
		}
	}

	// collapse when time is up.
	if best == nil {
		return end_at_edge(game, depth, max, preval)
	}

	flip := 1
	flag := AlphaFlag
	value := alpha
	m := make([]*chess.Move, 0, len(move_sorting))
	for move := range move_sorting { // converts map to list for sorting and return
		m = append(m, move)
	}
	if !max {
		flip = -1
		flag = BetaFlag
		value = beta
	}

	// if a break ends, keep the uninvestigated moves but move them to end of list
	if len(m) != len(moves) {
		for _, move := range moves[len(m)-1:] {
			move_sorting[move] = flip * -1 * 10000
		}
	}

	// sort moves by how good they were
	sort.Slice(m, func(i, j int) bool { return flip*move_sorting[m[i]] > flip*move_sorting[m[j]] })

	// save this in the transposition table (ignores if time over)
	write_hash(game.Position(), zobrist(game.Position().Board(), max), depth, flag, value, best, m)
	print_minmax_root_end(root)
	return best, eval, history, false
}

func quiescence_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int, moves []*chess.Move) (best *chess.Move, eval int, history [mem_size]string, ignore bool) {
	if max {
		eval = preval
		for _, move := range moves {

			// create a new game and simulate the move
			post := game.Clone()
			post.Move(move)

			state_eval := evaluate_position(game, post, preval, move)

			_, tempeval, temphistory, ignore := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)

			if ignore {
				continue
			}

			// checkmate for black
			if tempeval > eval {
				eval = tempeval
				best = move
				temphistory[DEPTH-depth] = move.String() + "q"
				history = temphistory
			}

			if tempeval > alpha {
				alpha = tempeval
			}

			// there exists a preferrable path elsewhere that will always be better for me
			// my opponent has an option in this branch that i can't avoid, not looking anymore
			if alpha >= beta {
				break
			}
		}
	} else {
		eval = preval
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval, temphistory, ignore := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if ignore {
				continue
			}

			if tempeval < eval {
				eval = tempeval
				best = move
				temphistory[DEPTH-depth] = move.String() + "q"
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
	if best == nil {
		return end_at_edge(game, depth, max, preval)
	}
	if !check_time_up() {
		if max {
			write_hash(game.Position(), zobrist(game.Position().Board(), max), depth, AlphaQFlag, alpha, best, moves)
		} else {
			write_hash(game.Position(), zobrist(game.Position().Board(), max), depth, BetaQFlag, beta, best, moves)
		}
	}
	return best, eval, history, false
}

func check_time_up() bool {
	if !DO_ITERATIVE_DEEPENING || !DO_STRICT_TIMING {
		return false
	}
	return delay.Sub(time.Now()) < 0
}

func end_at_edge(game *chess.Game, depth int, max bool, preval int) (best *chess.Move, eval int, history [mem_size]string, ignore bool) {
	if check_time_up() {
		history[DEPTH-depth] = "null"
		return nil, 0, history, true
	}
	history[DEPTH-depth] = "edge"
	write_hash(game.Position(), zobrist(game.Position().Board(), max), depth, EdgeFlag, preval, nil, nil)
	return nil, preval, history, false // history is blank
}

// -------------------------

