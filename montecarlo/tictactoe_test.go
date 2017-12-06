package montecarlo_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"gitlab.com/jac32/Settlers-of-Catan/game/montecarlo"
)

type cell int8

const (
	empty   = cell(0)
	playerX = cell(1)
	playerO = cell(2)
)

type gameState struct {
	turn  bool
	board [][]cell
}

func (state gameState) String() string {
	s := ""
	for row := 0; row < 3; row++ {
		s += "\t"
		for col := 0; col < 3; col++ {
			if state.board[row][col] == playerX {
				s += " X"
			} else if state.board[row][col] == playerO {
				s += " O"
			} else {
				s += " -"
			}
		}
		s += "\n"
	}
	return s
}

// isEnd returns true if there is a winner, along with either constant "playerOne",
// or "playerTwo". If nobody won, then false is returned with the constant "empty".
func (state gameState) isEnd() (bool, cell) {
	//check for each diagonal case
	if win, player := state.checkDiagonals(); win {
		return true, player
	}
	for i := 0; i < 3; i++ {
		//check for columns
		if win, player := state.checkCol(i); win {
			return true, player
		}
		if win, player := state.checkRow(i); win {
			return true, player
		}
	}
	//if there is no winner, check for a draw
	if !state.hasEmpty() {
		return true, empty
	}
	return false, empty
}

func (state gameState) hasEmpty() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if state.board[i][j] == empty {
				return true
			}
		}
	}
	return false
}

func (state gameState) checkRow(n int) (bool, cell) {
	xCount := 0
	oCount := 0
	for i := 0; i < 3; i++ {
		switch state.board[n][i] {
		case playerX:
			xCount++
		case playerO:
			oCount++
		default:
			//there is no way this row could be a winner
			return false, empty
		}
		if xCount == 3 {
			return true, playerX
		} else if oCount == 3 {
			return true, playerO
		}
	}
	return false, empty
}

func (state gameState) checkCol(n int) (bool, cell) {
	xCount := 0
	oCount := 0
	for i := 0; i < 3; i++ {
		switch state.board[i][n] {
		case playerX:
			xCount++
		case playerO:
			oCount++
		default:
			//there is no way this column could be a winner
			return false, empty
		}
		if xCount == 3 {
			return true, playerX
		} else if oCount == 3 {
			return true, playerO
		}
	}
	return false, empty
}

func (state gameState) checkDiagonals() (bool, cell) {
	if (state.board[0][0] == playerX &&
		state.board[1][1] == playerX &&
		state.board[2][2] == playerX) ||
		(state.board[2][0] == playerX &&
			state.board[1][1] == playerX &&
			state.board[0][2] == playerX) {
		return true, playerX
	} else if (state.board[0][0] == playerO &&
		state.board[1][1] == playerO &&
		state.board[2][2] == playerO) ||
		(state.board[2][0] == playerO &&
			state.board[1][1] == playerO &&
			state.board[0][2] == playerO) {
		return true, playerO
	}
	return false, empty
}

var actions montecarlo.ActionSet

func makeActions() {
	actions = make(montecarlo.ActionSet)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			var x, y int
			x = i
			y = j
			actions[fmt.Sprintf("SET_X_%v_%v", x, y)] = func(state montecarlo.State) montecarlo.State {
				s, _ := state.(gameState)
				s.board[x][y] = playerX
				s.turn = !s.turn
				return s
			}
			actions[fmt.Sprintf("SET_O_%v_%v", x, y)] = func(state montecarlo.State) montecarlo.State {
				s, _ := state.(gameState)
				s.board[x][y] = playerO
				s.turn = !s.turn
				return s
			}
		}
	}
}

/*-------- IMPLEMENT montecarlo.State INTERFACE --------*/

func (state gameState) LegalActions() montecarlo.ActionSet {
	legalActions := make(montecarlo.ActionSet)
	//legal actions include those which place a naught or a cross on an empty cell
	//it must also be the turn of the naught or the cross respectively
	if won, _ := state.isEnd(); won {
		return legalActions
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if state.board[i][j] == empty {
				var index montecarlo.Key
				if state.turn {
					index = fmt.Sprintf("SET_X_%v_%v", i, j)
				} else {
					index = fmt.Sprintf("SET_O_%v_%v", i, j)
				}
				legalActions[index] = actions[index]
			}
		}
	}
	return legalActions
}

func (state gameState) Score(player uint) float64 {
	//for the AI it should favour playerO winning
	win, p := state.isEnd()
	if win && p == playerO {
		if player == 0 {
			return 0
		}
		return float64(1)
	} else if win && p == playerX {
		if player == 1 {
			return 0
		}
		return float64(1)
	} else if win {
		// a draw
		return float64(0.5)
	}
	return float64(0)
}

func (state gameState) Bias() float64 {
	// bias is added for moves which are more likely to win, but not necessarily
	// better
	return float64(0)
}

func (state gameState) Copy() montecarlo.State {
	newState := gameState{
		turn:  state.turn,
		board: make([][]cell, 3, 3),
	}
	for i := 0; i < 3; i++ {
		newState.board[i] = make([]cell, 3, 3)
		for j := 0; j < 3; j++ {
			newState.board[i][j] = state.board[i][j]
		}
	}
	return newState
}

func (state gameState) Player() uint {
	if state.turn == true {
		// player X
		return 0
	}
	// player O
	return 1
}

func (state gameState) Policy() montecarlo.Policy {
	return montecarlo.UCTPolicy{}
}

func initState() gameState {
	state := gameState{
		//it is the human's turn first
		turn:  true,
		board: make([][]cell, 3, 3),
	}
	for i := 0; i < 3; i++ {
		state.board[i] = make([]cell, 3, 3)
		for j := 0; j < 3; j++ {
			state.board[i][j] = empty
		}
	}
	return state
}

func (state gameState) readValidChoice() (int, int) {
	valid := false
	x := 0
	y := 0
	for !valid {
		fmt.Printf("enter a cell coordinate \"row,col\" (zero-based): ")
		fmt.Scanf("%d,%d", &x, &y)
		valid = x >= 0 && x < 3 && y >= 0 && y < 3 && state.board[x][y] == empty
		if !valid {
			fmt.Println("invalid cell entry")
		}
	}
	return x, y
}

// runs an example match between montecarlo AIs.
func TestExample(t *testing.T) {
	makeActions()
	state := initState()
	//coin flip to determine who goes first; Xs or Os
	rand.Seed(time.Now().UTC().UnixNano())
	flip := rand.Intn(2)
	if flip == 0 {
		state.turn = !state.turn
	}
	//play the game
	var end bool
	var winner cell
	for end, winner = state.isEnd(); !end; end, winner = state.isEnd() {
		if state.turn {
			/*
				x, y := state.readValidChoice()
				state.board[x][y] = playerX
				state.turn = !state.turn
			*/
			fmt.Println("X's turn...")
			ai, err := montecarlo.NewMultiplayerMCTS(2, state, actions)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
			index, action, err := ai.Search(1000, float64(1)/math.Sqrt2)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
			fmt.Printf("X: \"taking action: %v\"\n", index)
			state = ((*action)(state)).(gameState)
		} else {
			fmt.Println("O's turn...")
			ai, err := montecarlo.NewMultiplayerMCTS(2, state, actions)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
			index, action, err := ai.Search(1000, float64(1)/math.Sqrt2)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
			fmt.Printf("O: \"taking action: %v\"\n", index)
			state = ((*action)(state)).(gameState)
		}
		fmt.Printf("%v", state)
	}
	if winner == playerX {
		fmt.Println("Xs win")
	} else if winner == playerO {
		fmt.Println("Os win")
	} else {
		fmt.Println("draw")
	}
}
