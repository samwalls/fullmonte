package montecarlo

import (
	"fmt"
	"testing"

	log "github.com/Sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

/*-------- TEST INPUTS & SETUP --------*/

var tree1, tree2, tree3 Tree
var possibleActions map[Key]Action

type treeTestBoolVector struct {
	A      bool
	B      bool
	RESULT bool
}

//example state to act upon (a 2D map)
type treeTestState [][]treeTestBoolVector

// implementing the State interface...

func (state treeTestState) LegalActions() ActionSet {
	return make(map[Key]Action, 0)
}

func (state treeTestState) SelectAction(actions ActionSet) (Key, Action) {
	return "0", actions["0"]
}

func (state treeTestState) Score(player uint) float64 {
	return float64(0)
}

func (state treeTestState) Bias() float64 {
	//TODO
	return float64(0)
}

func (state treeTestState) Copy() State {
	//no reference types used, no actions necessary
	return state
}

func (state treeTestState) Player() uint {
	return 0
}

func (state treeTestState) Policy() Policy {
	return nil
}

func treeTestSetupActions() {
	possibleActions = make(map[Key]Action, 0)
	/* TEMPLATE
	possibleActions["ACTION_NAME"] = func(state State) State {
		s, _ := state.(treeTestState)
		//TODO
		return s
	}
	*/
	//for all possible coordinates
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			//The values created here are necessary as function closures capture
			//by reference, the reference to i and j would always point to the
			//value of 10 - no good!
			var x, y int
			x = i
			y = j
			//set the "A", or "B" fields to true or false
			possibleActions["SET_"+fmt.Sprintf("%v", x)+"_"+fmt.Sprintf("%v", y)+"_A_TRUE"] = func(state State) State {
				s, _ := state.(treeTestState)
				s[x][y].A = true
				return s
			}
			possibleActions["SET_"+fmt.Sprintf("%v", x)+"_"+fmt.Sprintf("%v", y)+"_A_FALSE"] = func(state State) State {
				s, _ := state.(treeTestState)
				s[x][y].A = false
				return s
			}
			possibleActions["SET_"+fmt.Sprintf("%v", x)+"_"+fmt.Sprintf("%v", y)+"_B_TRUE"] = func(state State) State {
				s, _ := state.(treeTestState)
				s[x][y].B = true
				return s
			}
			possibleActions["SET_"+fmt.Sprintf("%v", x)+"_"+fmt.Sprintf("%v", y)+"_B_FALSE"] = func(state State) State {
				s, _ := state.(treeTestState)
				s[x][y].B = false
				return s
			}
			//put AND of A, B in RESULT
			possibleActions["AND_"+fmt.Sprintf("%v", x)+"_"+fmt.Sprintf("%v", y)] = func(state State) State {
				s, _ := state.(treeTestState)
				s[x][y].RESULT = s[x][y].A && s[x][y].B
				return s
			}
			//put OR of A, B in RESULT
			possibleActions["OR_"+fmt.Sprintf("%v", x)+"_"+fmt.Sprintf("%v", y)] = func(state State) State {
				s, _ := state.(treeTestState)
				s[x][y].RESULT = s[x][y].A || s[x][y].B
				return s
			}
			//put XOR of A, B in RESULT
			possibleActions["XOR_"+fmt.Sprintf("%v", x)+"_"+fmt.Sprintf("%v", y)] = func(state State) State {
				s, _ := state.(treeTestState)
				s[x][y].RESULT = (s[x][y].A || s[x][y].B) && !(s[x][y].A && s[x][y].B)
				return s
			}
		}
	}
	/*
		for key, _ := range possibleActions {
			log.Debugf("created action: %v", key)
		}
	*/
}

func treeTestSetup() {
	log.SetLevel(log.DebugLevel)
	treeTestSetupActions()
	initialState := make(treeTestState, 10)
	for i := 0; i < 10; i++ {
		initialState[i] = make([]treeTestBoolVector, 10)
		for j := 0; j < 10; j++ {
			initialState[i][j].A = false
			initialState[i][j].B = false
			initialState[i][j].RESULT = false
		}
	}
	var err error
	tree1, err = NewTree(1, initialState, possibleActions)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

/*-------- TESTING --------*/

// helper function to make testing less noisy
func treeTestToState(t *testing.T, state State) treeTestState {
	testState, ok := state.(treeTestState)
	assert.True(t, ok, "type assertion failed")
	return testState
}

func TestNewTree(t *testing.T) {
	treeTestSetup()
	assert.Equal(t, 10, len(treeTestToState(t, tree1.Root().State)))
	for i := 0; i < 10; i++ {
		assert.Equal(t, 10, len(treeTestToState(t, tree1.Root().State)[i]))
		for j := 0; j < 10; j++ {
			//check the initial state
			state := treeTestToState(t, tree1.Root().State)[i][j]
			assert.False(t, state.A)
			assert.False(t, state.B)
			assert.False(t, state.RESULT)
			//check that the expected functions actually exist
			_, ok := possibleActions["SET_"+fmt.Sprintf("%v", i)+"_"+fmt.Sprintf("%v", j)+"_A_TRUE"]
			assert.True(t, ok)
			_, ok = possibleActions["SET_"+fmt.Sprintf("%v", i)+"_"+fmt.Sprintf("%v", j)+"_A_TRUE"]
			assert.True(t, ok)
			_, ok = possibleActions["SET_"+fmt.Sprintf("%v", i)+"_"+fmt.Sprintf("%v", j)+"_B_TRUE"]
			assert.True(t, ok)
			_, ok = possibleActions["SET_"+fmt.Sprintf("%v", i)+"_"+fmt.Sprintf("%v", j)+"_B_FALSE"]
			assert.True(t, ok)
			_, ok = possibleActions["AND_"+fmt.Sprintf("%v", i)+"_"+fmt.Sprintf("%v", j)]
			assert.True(t, ok)
			_, ok = possibleActions["XOR_"+fmt.Sprintf("%v", i)+"_"+fmt.Sprintf("%v", j)]
			assert.True(t, ok)
			_, ok = possibleActions["OR_"+fmt.Sprintf("%v", i)+"_"+fmt.Sprintf("%v", j)]
			assert.True(t, ok)
		}
	}
}

func TestSetBoolAction(t *testing.T) {
	treeTestSetup()
	state := tree1.Root().State
	assert.False(t, treeTestToState(t, state)[0][0].A)
	state = possibleActions["SET_0_0_A_TRUE"](state)
	assert.True(t, treeTestToState(t, state)[0][0].A)
	state = possibleActions["SET_0_0_A_FALSE"](state)
	assert.False(t, treeTestToState(t, state)[0][0].A)
	state = possibleActions["SET_0_0_A_TRUE"](state)
	assert.True(t, treeTestToState(t, state)[0][0].A)
	state = possibleActions["SET_0_0_A_FALSE"](state)
	assert.False(t, treeTestToState(t, state)[0][0].A)

	assert.False(t, treeTestToState(t, state)[0][0].B)
	state = possibleActions["SET_0_0_B_TRUE"](state)
	assert.True(t, treeTestToState(t, state)[0][0].B)
	state = possibleActions["SET_0_0_B_FALSE"](state)
	assert.False(t, treeTestToState(t, state)[0][0].B)
	state = possibleActions["SET_0_0_B_TRUE"](state)
	assert.True(t, treeTestToState(t, state)[0][0].B)
	state = possibleActions["SET_0_0_B_FALSE"](state)
	assert.False(t, treeTestToState(t, state)[0][0].B)
}

func TestAndAction(t *testing.T) {
	treeTestSetup()
	state := tree1.Root().State
	state = possibleActions["AND_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_A_TRUE"](state)
	state = possibleActions["AND_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_B_TRUE"](state)
	state = possibleActions["AND_0_0"](state)
	assert.True(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_A_FALSE"](state)
	state = possibleActions["AND_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
}

func TestOrAction(t *testing.T) {
	treeTestSetup()
	state := tree1.Root().State
	state = possibleActions["OR_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_A_TRUE"](state)
	state = possibleActions["OR_0_0"](state)
	assert.True(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_B_TRUE"](state)
	state = possibleActions["OR_0_0"](state)
	assert.True(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_A_FALSE"](state)
	state = possibleActions["OR_0_0"](state)
	assert.True(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_B_FALSE"](state)
	state = possibleActions["OR_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
}

func TestXorAction(t *testing.T) {
	treeTestSetup()
	state := tree1.Root().State
	state = possibleActions["XOR_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_A_TRUE"](state)
	state = possibleActions["XOR_0_0"](state)
	assert.True(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_B_TRUE"](state)
	state = possibleActions["XOR_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_A_FALSE"](state)
	state = possibleActions["XOR_0_0"](state)
	assert.True(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
	state = possibleActions["SET_0_0_B_FALSE"](state)
	state = possibleActions["XOR_0_0"](state)
	assert.False(t, treeTestToState(t, state)[0][0].RESULT, "wrong result")
}
