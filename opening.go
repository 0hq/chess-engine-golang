package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/notnil/chess"
	"github.com/notnil/chess/opening"
)

func testopening(){
	// g := chess.NewGame()
    g := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	g.MoveStr("e2e4")
	// g.MoveStr("e6")
	opening := true
	for opening {
		move := get_opening(g, 0)
		if move == nil {
			opening = false
		}
		g.Move(move)
		fmt.Println(g.Position().Board().Draw())
	}
	

	
	

}

func get_opening(g *chess.Game, retries int) *chess.Move {
	book := opening.NewBookECO()
	moves := g.Moves()
	if len(moves) == 0 && g.FEN() == "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" {
		var gx *chess.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
		gx.MoveStr("e2e4")
		fmt.Println(gx.Moves())
		return gx.Moves()[0]
	}
	o := book.Find(moves) // find current opening
	if o == nil {
		return nil
	}
	fmt.Println("\nFrom:", o.Title())
	// fmt.Println(g.Moves())
	p := book.Possible(g.Moves()) // all openings available
	if len(p) > 0 {
		r := p[rand.Intn(len(p))] // random opening available
		fmt.Println("To:", r.Title())
		fmt.Println(r.PGN())
		// pgn, err := chess.PGN(bytes.NewBufferString(r.PGN()))
		// if err != nil {
		// 	panic(err)
		// }
		split := strings.Split(r.PGN(), " ")
		if len(split) <= len(g.Moves()) {
			if len(p) == 1 || retries > 3 {
				return nil
			}
			return get_opening(g, retries + 1)
		}
		gx := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
		for _, s := range split {
			gx.MoveStr(s)
		}
		ms := gx.Moves()
		// fmt.Println(split)
		m := ms[len(g.Moves())]
		fmt.Println(m)
		return m
	}
	return nil
}

func get_opening_uci(g *chess.Game, retries int) string {
	book := opening.NewBookECO()
	moves := g.Moves()
	if len(moves) == 0 {
		return "e2e4"
	}
	o := book.Find(moves) // find current opening
	fmt.Println("\nFrom:", o.Title())
	// fmt.Println(g.Moves())
	p := book.Possible(g.Moves()) // all openings available
	if len(p) > 0 {
		r := p[rand.Intn(len(p))] // random opening available
		fmt.Println("To:", r.Title())
		fmt.Println(r.PGN())
		split := strings.Split(r.PGN(), " ")
		if len(split) <= len(g.Moves()) {
			if len(p) == 1 || retries > 3 {
				return ""
			}
			return get_opening_uci(g, retries + 1)
		}
		// fmt.Println(split)
		m := split[len(g.Moves())]
		fmt.Println(m)
		return m
	}
	return ""
}