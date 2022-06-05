package main

import (
	"sort"

	"github.com/notnil/chess"
)

// ------ constants -------

func PieceValue(p chess.PieceType) int {
	types := chess.PieceTypes()
	switch p {
	case types[0]:
		return 20000
	case types[1]:
		return 900
	case types[2]:
		return 500
	case types[3]:
		return 330
	case types[4]:
		return 320
	case types[5]:
		return 100
	}
	return -1
}

var pos_p = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{50, 50, 50, 50, 50, 50, 50, 50},
	{10, 10, 20, 30, 30, 20, 10, 10},
	{5, 5, 10, 25, 25, 10, 5, 5},
	{0, 0, 0, 20, 20, 0, 0, 0},
	{5, -5, -10, 0, 0, -10, -5, 5},
	{5, 10, 10, -20, -20, 10, 10, 5},
	{0, 0, 0, 0, 0, 0, 0, 0},
}
var pos_n = [8][8]int{
	{-50, -40, -30, -30, -30, -30, -40, -50},
	{-40, -20, 0, 0, 0, 0, -20, -40},
	{-30, 0, 10, 15, 15, 10, 0, -30},
	{-30, 5, 15, 20, 20, 15, 5, -30},
	{-30, 0, 15, 20, 20, 15, 0, -30},
	{-30, 5, 10, 15, 15, 10, 5, -30},
	{-40, -20, 0, 5, 5, 0, -20, -40},
	{-50, -40, -30, -30, -30, -30, -40, -50},
}
var pos_b = [8][8]int{
	{-20, -10, -10, -10, -10, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 10, 10, 5, 0, -10},
	{-10, 5, 5, 10, 10, 5, 5, -10},
	{-10, 0, 10, 10, 10, 10, 0, -10},
	{-10, 10, 10, 10, 10, 10, 10, -10},
	{-10, 5, 0, 0, 0, 0, 5, -10},
	{-20, -10, -10, -10, -10, -10, -10, -20},
}
var pos_r = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{5, 10, 10, 10, 10, 10, 10, 5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{0, 0, 0, 5, 5, 0, 0, 0},
}
var pos_q = [8][8]int{
	{-20, -10, -10, -5, -5, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 5, 5, 5, 0, -10},
	{-5, 0, 5, 5, 5, 5, 0, -5},
	{0, 0, 5, 5, 5, 5, 0, -5},
	{-10, 5, 5, 5, 5, 5, 0, -10},
	{-10, 0, 5, 0, 0, 0, 0, -10},
	{-20, -10, -10, -5, -5, -10, -10, -20},
}
var pos_k = [8][8]int{
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-20, -30, -30, -40, -40, -30, -30, -20},
	{-10, -20, -20, -20, -20, -20, -20, -10},
	{20, 20, 0, 0, 0, 0, 20, 20},
	{20, 30, 10, 0, 0, 10, 30, 20},
}
var pos_k_endgame = [8][8]int{
	{-50, -40, -30, -20, -20, -30, -40, -50},
	{-30, -20, -10, 0, 0, -10, -20, -30},
	{-30, -10, 20, 30, 30, 20, -10, -30},
	{-30, -10, 30, 40, 40, 30, -10, -30},
	{-30, -10, 30, 40, 40, 30, -10, -30},
	{-30, -10, 20, 30, 30, 20, -10, -30},
	{-30, -30, 0, 0, 0, 0, -30, -30},
	{-50, -30, -30, -30, -30, -30, -30, -50},
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

	if move.HasTag(chess.Capture) {
		// fmt.Println(PieceValue(game.Position().Board().Piece(move.S2()).Type()))
		eval += flip * PieceValue(pre.Position().Board().Piece(move.S2()).Type())
	}

	return eval
}

func get_quiescence_moves(game *chess.Game, moves []*chess.Move) []*chess.Move {
	funcVar := func(move *chess.Move) bool {
		if move.HasTag(chess.Capture) {
			return true
		}
		// perhaps include check moves?
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
	eval += to - from

	return
}

func evaluate_quiescence_move(game *chess.Game, move *chess.Move) int {
	move_type := game.Position().Board().Piece(move.S1()).Type()
	return PieceValue(game.Position().Board().Piece(move.S2()).Type()) - PieceValue(move_type)
}

// this should be static, not relative
func update_evaluation(game *chess.Game, pre *chess.Game, move *chess.Move) {
	position_eval = evaluate_position(pre, game, position_eval, move)
}
