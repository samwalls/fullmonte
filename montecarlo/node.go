package montecarlo

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Key is the key type used to map to child nodes (actions)
type Key interface{}

// Node contains information about a singular, immutable game state along with
// information pertaining to MCTS.
type Node struct {
	score      []float64
	numPlayers uint
	State      State
	visits     int64
	parent     *Node
	// The string index in children is analogous to the string of the action in
	// the montecarlo.Tree's possibleActions, that would lead to the child state
	// from this node.
	children map[Key]*Node
	policy   Policy
}

// NewNode creates a fully formed MCTS tree node, containing a nil state however.
// numPlayers is the number of players in the MCTS
func NewNode(numPlayers uint) (Node, error) {
	n := Node{
		score:      make([]float64, numPlayers),
		numPlayers: numPlayers,
		parent:     nil,
		children:   make(map[Key]*Node, 0),
		policy:     UCTPolicy{},
	}
	if numPlayers <= 0 {
		return n, ZeroPlayerCount(n)
	}
	return n, nil
}

func (node Node) String() string {
	out := "Node {\n"
	out += fmt.Sprintf("\tstate: %v\n", node.State)
	out += fmt.Sprintf("\tnumPlayers: %v\n", node.NumPlayers())
	out += fmt.Sprintf("\tscoreVector: %v\n", node.ScoreVector())
	out += fmt.Sprintf("\tcurrentPlayer: %v\n", node.Player())
	out += "\tchildren: [\n"
	for k := range node.children {
		out += fmt.Sprintf("\t\t%v,\n", k)
	}
	out += "\t]\n"
	out += "}"
	return out
}

// Copy returns a deep copy of this node
func (node Node) Copy() *Node {
	// will not throw any error since we're already using a valid player count
	cpy, _ := NewNode(node.NumPlayers())
	cpy.children = make(map[Key]*Node)
	// add the nodes of this tree into the copy
	// will not throw an error since the player counts are the same
	_ = cpy.Merge(node)
	return &cpy
}

// Merge two nodes and all their children: add all nodes from other into this
// node's tree of children. If both trees have the same node, then their Score
// and Visit values are added.
func (node *Node) Merge(other Node) error {
	//TODO make a version of Merge that does not create side-effects
	// check for any errors that mean the node cannot be merged
	err := node.mergeGetErrors(other)
	if err != nil {
		return err
	}
	// add/merge node values
	for i := uint(0); i < node.NumPlayers(); i++ {
		node.score[i] += other.score[i]
	}
	node.visits += other.Visits()
	if node.State == nil && other.State != nil {
		node.State = other.State.Copy()
	}
	// Add child nodes from other to this node.
	// If this node doesn't contain a child from other, it is copied.
	// All nodes in common are recursively merged (depth-first).
	for k, otherChild := range other.children {
		if otherChild == nil {
			continue
		}
		if _, ok := node.children[k]; !ok {
			// if a child with that key does not exist on this node, make it
			otherCopy := otherChild.Copy()
			// copies work from the node down (parent is culled out)
			node.SetChild(k, otherCopy)
		} else {
			node.GetChild(k).Merge(*otherChild)
		}
	}
	return nil
}

// Generates any errors associated with merging with the node "other".
func (node *Node) mergeGetErrors(other Node) error {
	// player count needs to be consistent
	if other.NumPlayers() != node.NumPlayers() {
		return MergeDifferingPlayerCount{
			node.NumPlayers(),
			other.NumPlayers(),
		}
	}
	// player at this state needs to be consistent
	if node.State != nil && other.State != nil && node.Player() != other.Player() {
		return MergePlayerIndexMismatch{
			node.Player(),
			other.Player(),
		}
	}
	// state needs to be consistent (if the first two errors do not
	// adequately capture this already)
	//if node.State != other.State {
	//	return MergeStateMismatch{
	//		node.State,
	//		other.State,
	//	}
	//}
	return nil
}

// UpperConfidenceBound (UCB) describes the upper end of the confidence bound
// (a range for which a certain percentage of probabilities are correct) in
// terms of nodes that look promising to exploit (are proven to have a good
// score) vs. nodes that look promising to explore (have not been explored very
// deeply).
//
// Of a node's children, the most likely to get the best score is the one with
// the highest UCB value.
//
// The UCB can be tweaked to favour exploration when the exploration parameter
// is high. If the exploration parameter is zero (the value will be clamped to
// [0, infinity]) no voluntary exploration will be favoured.
func (node Node) UpperConfidenceBound(explorationParameter float64, player uint) float64 {
	var expl float64
	if explorationParameter < 0 {
		expl = 0
	} else {
		expl = explorationParameter
	}
	nVisits := float64(node.Visits())
	//inflate UCB for nodes which are a direct child of the root node
	if node.IsRoot() || nVisits <= 0 || nVisits == 1 && !node.IsRoot() && node.Parent().IsRoot() {
		// all nodes with no visits will have an equal chance (positive infinity)
		return math.Inf(1)
	}
	nScore := node.Score(player)
	pVisits := float64(node.Parent().Visits())
	return (nScore / nVisits) + expl*math.Sqrt(float64(2)*math.Log(pVisits)/nVisits)
}

// selectBestChild returns the string for an action with the highest rated
// confidence, as well as the resulting state (Node). The string is a key to
// the MCTS tree's list of possible actions.
// If the node has no children, then the empty string is returned along with the
// node itself.
func (node Node) selectBestChild(explorationParam float64) (Key, *Node) {
	maxUCB := math.Inf(-1)
	maxIndex := interface{}(nil)
	if node.IsLeaf() {
		return "", &node
	}
	epsilon := 0.000001
	maxima := make(map[float64][]Key)
	//find the highest upper-confidence-bound in this node's children
	for i, n := range node.children {
		//we calculate the upper confidence bound for the child's player itself;
		//not the root node's player - this is because we imagine that each
		//player will try to maximise their own reward (Browne et al. page 10 -
		//"Multiplayer MCTS").
		ucb := n.UpperConfidenceBound(explorationParam, node.Player())
		// add selection bias for nodes containing states that specifiy it
		bias := float64(0)
		if n.State != nil {
			bias = n.State.Bias()
		}
		if ucb+bias >= maxUCB {
			//compare floats within range of epsilon
			if (maxUCB - ucb) <= epsilon {
				maxUCB = ucb
				maxIndex = i
				maxima[maxUCB] = make([]Key, 0)
			}
			maxima[maxUCB] = append(maxima[maxUCB], maxIndex)
		}
	}
	//if there is no true maximum, pick a random one
	if len(maxima[maxUCB]) > 1 {
		n := len(maxima[maxUCB])
		rand.Seed(time.Now().UTC().UnixNano())
		target := rand.Intn(n)
		i := 0
		for k := range maxima[maxUCB] {
			if i == target {
				return maxima[maxUCB][k], node.children[maxima[maxUCB][k]]
			}
			i++
		}
	}
	return maxIndex, node.children[maxIndex]
}

// Parent returns the parent of this node.
func (node Node) Parent() *Node {
	return node.parent
}

// SetChild sets the child of this node (at the specified index) to the passed
// child.
func (node *Node) SetChild(index Key, child *Node) {
	child.parent = node
	node.children[index] = child
}

// RemoveChild removes the child with the specified index from this node's set
// of children (if it exists).
func (node *Node) RemoveChild(index Key) {
	_, ok := node.children[index]
	if ok {
		node.children[index].parent = nil
		delete(node.children, index)
	}
}

// GetChild returns the child of the specified index from this node's set of
// children.
func (node Node) GetChild(index Key) *Node {
	return node.children[index]
}

// IsLeaf returns true if the called-upon node is a leaf node in the tree false
// otherwise.
func (node Node) IsLeaf() bool {
	return node.children == nil || len(node.children) == 0
}

// IsTerminal conceptually differs from IsLeaf in that a node will be called
// "terminal" if it's domain state is terminal (end of the game), whereas IsLeaf
// returns true if it is merely the node's position in the tree that is terminal.
func (node Node) IsTerminal() bool {
	return node.State == nil || len(node.State.LegalActions()) == 0
}

// IsRoot returns true if the called-upon node has no parent (and is in fact a
// root), false otherwise.
func (node Node) IsRoot() bool {
	return node.parent == nil
}

// Visits returns the number of visits the node has had.
func (node Node) Visits() int64 {
	return node.visits
}

// AddVisit increments the number of visits of this node.
func (node *Node) AddVisit() {
	node.visits++
}

// IsExhausted returns true if all possible actions have been created for this
// node. If the node happens to have a nil state, then true is also returned.
func (node Node) IsExhausted() bool {
	// if any child is not exhausted then this node is not either
	if node.State == nil {
		return true
	}
	for k := range node.State.LegalActions() {
		if _, ok := node.children[k]; !ok {
			return false
		}
	}
	return true
}

// NumPlayers gets the number of players participating in MCTS
func (node Node) NumPlayers() uint {
	return node.numPlayers
}

// Player gets the current player in this node
func (node Node) Player() uint {
	if node.State == nil {
		return 0
	}
	return node.State.Player()
}

// ScoreVector gets the array containing the scores of all players for this node.
func (node Node) ScoreVector() []float64 {
	return node.score
}

// Score gets the score of the specified player
func (node Node) Score(player uint) float64 {
	return node.score[player]
}

// SetScore sets the score of the specified player
func (node *Node) SetScore(player uint, score float64) {
	node.score[player] = score
}

// Policy returns the policy used by this node
func (node Node) Policy() Policy {
	return node.policy
}
