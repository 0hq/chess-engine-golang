package main

import (
	"fmt"
	"math"
	"sort"

	"github.com/notnil/chess"
)

func quiescence_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int, move_gen []*chess.Move) (best *chess.Move, eval int, history []*chess.Move) {
	moves := get_quiescence_moves(game, move_gen)

	if len(moves) == 0 {
		write_hash(zobrist(game.Position().Board(), max), depth, "EDGE", preval, nil, nil, game.Position())
		return nil, preval, history // history is blank here
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
				history = append(temphistory, move)
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
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
				history = append(temphistory, move)
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

func minimax_hashing(game *chess.Game, depth int, alpha int, beta int, max bool, preval int) (best *chess.Move, eval int, history []*chess.Move) {

	if depth < MAX_QUIESCENCE {
		write_hash(zobrist(game.Position().Board(), max), depth, "EDGE", preval, nil, nil, game.Position())
		return nil, preval, history // history is blank
	}

	var moves []*chess.Move
	flag, hashscore, hashbest, hashmoves := read_hash(zobrist(game.Position().Board(), max), depth, alpha, beta)

	if flag == 1 {
		history = append(history, hashbest)
		return hashbest, hashscore, history
	} else if flag == 2 {
		moves = hashmoves
	} else {
		moves = game.ValidMoves()
	}

	explored++
	explored_depth[DEPTH-depth]++

	// makes sure we don't run quiescence move pruning on empty
	if len(moves) == 0 {
		write_hash(zobrist(game.Position().Board(), max), depth, "EDGE", preval, nil, nil, game.Position())
		return nil, preval, history // history is blank here
	}

	if depth <= 0 {
		return quiescence_hashing(game, depth, alpha, beta, max, preval, moves)
	}

	root := depth == DEPTH
	if flag == 2 {
		moves = move_order_hashing(game, moves, hashbest)
	} else {
		moves = move_order(game, moves)
	}

	if root {
		fmt.Println("MOVE ORDER:")
		fmt.Println("HASH RETURN:", hashbest, hashmoves, flag, hashscore)
	}

	if max {
		eval = -1 * math.MaxInt
		for _, move := range moves {
			post := game.Clone()
			post.Move(move)
			state_eval := evaluate_position(game, post, preval, move)
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval > eval {
				eval = tempeval
				best = move
				history = append(temphistory, move)
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
			_, tempeval, temphistory := minimax_hashing(post, depth-1, alpha, beta, !max, state_eval)
			if tempeval < eval {
				eval = tempeval
				best = move
				history = append(temphistory, move)
				if root {
					fmt.Println()
					fmt.Println(move, tempeval, beta)
					fmt.Println(history)
				}
			}
			if tempeval < beta {
				beta = tempeval
			}
			if alpha >= beta {
				break
			}
			if root {
				fmt.Print("x")
			}
		}
	}
	if root {
		fmt.Print("x")
	}

	if max {
		write_hash(zobrist(game.Position().Board(), max), depth, "ALPHA", alpha, best, moves, game.Position())
	} else {
		write_hash(zobrist(game.Position().Board(), max), depth, "BETA", beta, best, moves, game.Position())
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

func move_order_hashing(game *chess.Game, moves []*chess.Move, best *chess.Move) []*chess.Move {
	// const len int = len(moves)
	// fmt.Println(moves)
	evaluated := make(map[*chess.Move]int)
	for _, move := range moves {
		evaluated[move] = evaluate_move(game, move)
	}

	keys := make([]*chess.Move, 0, len(evaluated))

	for key := range evaluated {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return evaluated[keys[i]] > evaluated[keys[j]] })

	if best != nil {
		var t []*chess.Move
		t = append(t, best)
		keys = append(t, keys...)
	}

	// fmt.Println(keys)
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
