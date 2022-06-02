package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/notnil/chess"
)

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
