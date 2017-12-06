package montecarlo

import (
	"fmt"
	"math/rand"
	"time"
)

// UCTPolicy is based on the UCT algorithm outlined by (Browne et al. 2012: A
// Survey of Monte Carlo Tree Search Methods - IEEE transactions on
// computational intelligence and AI in games, vol. 4, no. 1).
type UCTPolicy struct{}

/******** IMPLEMENT Policy ********/

// Select selects the child with the highest UCB
func (p UCTPolicy) Select(node *Node, explorationParam float64) *Node {
	//_, n := node.selectBestLeaf(expl)
	n := node
	for n != nil && (n.IsRoot() || !n.IsTerminal()) {
		if !n.IsExhausted() {
			return p.Expand(n, explorationParam)
		}
		_, n = n.selectBestChild(explorationParam)
	}
	return n
}

// Expand adds all actions that are legal from the passed node, and selects one
// to simulate/playout.
func (p UCTPolicy) Expand(node *Node, explorationParam float64) *Node {
	legalActions := node.State.LegalActions()
	untried := make(map[Key]Action)
	// add children if they haven't already been added
	for k, action := range legalActions {
		if node.GetChild(k) == nil {
			untried[k] = action
		}
	}
	// choose an action from the set of untried actions
	index, action := randomAction(untried)
	if action == nil {
		return node
	}
	// add a child to the node
	n, err := NewNode(node.NumPlayers())
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	n.State = (*action)(node.State.Copy())
	n.policy = n.State.Policy()
	node.SetChild(index, &n)
	return &n
}

// Simulate by stochastically selecting legal moves until the end of the
// simulation is reached.
func (p UCTPolicy) Simulate(node *Node) float64 {
	score := float64(0)
	n := 1
	//take the average of n simulations?
	for i := 0; i < n; i++ {
		state := node.State.Copy()
		// take random actions ad nauseum
		for {
			legalActions := state.LegalActions()
			if len(legalActions) <= 0 {
				break
			}
			_, action := randomAction(legalActions)
			state = (*action)(state)
		}
		score += state.Score(state.Player())
	}
	return score / float64(n)
}

// Backpropagate propagates the same score up the tree until the root is
// reached; the number of visits is also incremented at each node on the way.
func (p UCTPolicy) Backpropagate(node *Node, score float64) {
	player := node.Player()
	n := node
	for n != nil {
		n.SetScore(player, n.Score(player)+score)
		n.AddVisit()
		n = n.Parent()
	}
}

// randomAction returns a random string, action pair from a map of actions
func randomAction(actions map[Key]Action) (Key, *Action) {
	numActions := len(actions)
	target := 0
	if numActions > 0 {
		rand.Seed(time.Now().UTC().UnixNano())
		target = rand.Intn(numActions)
	}
	i := 0
	for k, v := range actions {
		if i == target {
			return k, &v
		}
		i++
	}
	return "", nil
}
