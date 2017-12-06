package montecarlo

import "sync"

// ActionSet is a map from string to action
type ActionSet map[Key]Action

// Action is a simple transition from State to State, domain actions are defined
// in a montecarlo tree search by the possibleActions parameter of NewTree.
type Action func(state State) State

// State is an interface that defines what a type must be able to do, so that
// it can be used in MCTS.
type State interface {
	// LegalActions is part of the expansion stage, all legal actions
	// from the current state should be returned. The string indices are
	// analogous to the indices used in the definition of all possible actions
	// (see NewTree).
	LegalActions() ActionSet
	// Score returns the score of a node for the given player - according to the
	// game rules
	Score(player uint) float64
	// Bias also part of simulation phase, create a score bias based on domain
	// knowledge
	Bias() float64
	// Copy is necessary for storing and performing actions on states without
	// modifying the original state itself
	Copy() State
	// Player gives the current player number
	Player() uint
	// Policy gives the policy type to be used given the current state
	Policy() Policy
}

/*-------- TreeSearcher DEFAULT IMPLEMENTATION --------*/

// MultiplayerMCTS encapsulates the information required for a basic MCTS
// implementation, including: its search tree; and its policy for operating on
// the tree.
type MultiplayerMCTS struct {
	tree   Tree
	policy Policy
}

// NewMultiplayerMCTS creates a new context from which to run a basic MCTS.
func NewMultiplayerMCTS(numPlayers uint, init State, actions map[Key]Action) (MultiplayerMCTS, error) {
	t, err := NewTree(numPlayers, init.Copy(), actions)
	mcts := MultiplayerMCTS{
		tree: t,
	}
	return mcts, err
}

// Search via MCTS, in a single-threaded manner, for the best action to take.
// Returns the index of the best action to take, as well as the action itself
// (according to the list of possible actions).
func (mcts MultiplayerMCTS) Search(level int64, expl float64) (Key, *Action, error) {
	for i := int64(0); i < level; i++ {
		root := mcts.tree.Root()
		node := root.Policy().Select(&root, expl)
		node.Policy().Backpropagate(node, node.Policy().Simulate(node))
		//log.Infof("finished %vth simulation", i)
	}
	/*
		// debugging output
		for k, c := range mcts.tree.Root().children {
			fmt.Printf("{%v: %v/%v | current player: %v | UCB %v}\n",
				k,
				c.ScoreVector(),
				c.Visits(),
				c.Player(),
				c.UpperConfidenceBound(expl, c.Player()),
			)
			for k2, c2 := range c.children {
				fmt.Printf("\t{ %v: %v/%v | current player: %v | UCB %v}\n",
					k2,
					c2.ScoreVector(),
					c2.Visits(),
					c2.Player(),
					c2.UpperConfidenceBound(expl, c2.Player()),
				)
			}
		}
	*/
	// maximise exploitation over exploration by setting the exploration parameter
	// to 0
	key, _ := mcts.tree.root.selectBestChild(0)
	action := mcts.tree.PossibleActions()[key]
	return key, &action, nil
}

// RootParallelSearch searches via MCTS, in a root-parallel manner, for the best
// action to take. Returns the key of the best action to take, as well as the
// action itself (according to the list of possible actions).
// TODO fix this
func (mcts MultiplayerMCTS) RootParallelSearch(numThreads int, level int64, expl float64) (Key, *Action, error) {
	// an output channel of the final trees each thread produces
	trees := make(chan *Tree, numThreads)
	var counter sync.WaitGroup
	counter.Add(numThreads)
	for threadno := 0; threadno < numThreads; threadno++ {
		go func() {
			defer counter.Done()
			// create a separate copy of the initial tree for this thread
			tree := mcts.tree.Copy()
			for i := int64(0); i < level; i++ {
				root := tree.Root()
				node := root.Policy().Select(&root, expl)
				node.Policy().Backpropagate(node, node.Policy().Simulate(node))
			}
			trees <- tree
		}()
	}
	// merge all created trees
	go func() {
		for t := range trees {
			mcts.tree.Merge(*t)
		}
	}()
	// wait for all searches to finish
	counter.Wait()
	key, _ := mcts.tree.root.selectBestChild(0)
	action := mcts.tree.PossibleActions()[key]
	return key, &action, nil
}
