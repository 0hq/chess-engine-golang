package main

import (
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
)

// func init_stockfish() *uci.Engine {
// 	eng, err := uci.New("stockfish")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer eng.Close()
// 	// initialize uci with new game
// 	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
// 		panic(err)
// 	}
// 	return eng
// }

func stockfish(game *chess.Game, eng *uci.Engine) *chess.Move {
	cmdPos := uci.CmdPosition{Position: game.Position()}
	cmdGo := uci.CmdGo{MoveTime: time.Second / 100}
	if err := eng.Run(cmdPos, cmdGo); err != nil {
		panic(err)
	}
	return eng.SearchResults().BestMove
}

func random_move_engine(game *chess.Game) *chess.Move { // about as good as stockfish ofc
	moves := game.ValidMoves()
	return moves[0]
}
