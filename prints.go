package main

import (
	"fmt"
	"time"

	"github.com/notnil/chess"
)

func print_root_move_1(root bool, game *chess.Game, move *chess.Move, tempeval int, cap int, history [mem_size]string) {
	if VERBOSE_FLAG < 2 || !root {
		return
	}
	fmt.Println("\nNew best root move:", move)
	fmt.Println("Evaluation:", tempeval, "Prev eval (forced beta/alpha):", cap)
	fmt.Println("Move path:", history)
	// fmt.Println(game.Position().Board().Draw())
}

func print_root_move_2(root bool) {
	if VERBOSE_FLAG < 2 || !root {
		return
	}
	fmt.Print("x")
}

func print_minmax_root_end(root bool) {
	if VERBOSE_FLAG < 2 || root {
		return
	}
	fmt.Print("\n")
	if check_time_up() {
		fmt.Println("\n -- Returned early because time was up... -- ")
	}
}

func print_iter_1(delay time.Time) {
	if VERBOSE_FLAG < 1 {
		return
	}
	fmt.Println("\n -- Searching deeper --")
	fmt.Println("Depth:", DEPTH)
	fmt.Println("Time left:", delay.Sub(time.Now()), "\n")
}

func print_iter_11(output *chess.Move, eval int, history [mem_size]string) {
	if VERBOSE_FLAG < 1 {
		return
	}
	fmt.Println("\n", output)
	fmt.Println("line:", history)
	fmt.Println("evaluation:", eval)
}

func print_iter_2() {
	if VERBOSE_FLAG < 1 {
		return
	}

	fmt.Println("\nTotal nodes explored", explored)
	fmt.Println("# nodes at depth", explored_depth)
	fmt.Println("Total hashes used", hash_count)
	fmt.Println("Hashes written", hash_write_count)
	fmt.Println("Hash types (edge, alpha, beta)", hash_count_list)
}

func print_turn_complete(game *chess.Game, move *chess.Move, start time.Time) {
	fmt.Println("\nMove chosen:", move)
	end := time.Now()
	fmt.Println(game.Position().Board().Draw())
	fmt.Println("Time elapsed", end.Sub(start))
	fmt.Println(`[SetUp "1"]`)
	fmt.Print(`[FEN "`, start_pos, `"]`, "\n")
	fmt.Println(game)
	// fmt.Println(game.Position())
	// fmt.Println(game)
}

func print_game_over(game *chess.Game) {
	fmt.Printf("\n\n ----- Game completed. %s by %s. ------\n\n", game.Outcome(), game.Method())
	fmt.Println(`[SetUp "1"]`)
	fmt.Print(`[FEN "`, start_pos, `"]`, "\n")
	fmt.Println(game)

}
