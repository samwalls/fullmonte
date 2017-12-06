package montecarlo

// DefaultPolicy describes how a montecarlo tree search should simulate a
// playout from a particular node. The default policy can be thought of as a
// policy that describes what would happen were MCTS not being used at all.
type DefaultPolicy interface {
	// Simulate returns the score of a simulation from the passed node.
	Simulate(node *Node) float64
}

// TreePolicy describes how a montecarlo tree search should select which nodes
// to visit.
type TreePolicy interface {
	// Select picks a node simulate from.
	Select(node *Node, explorationParam float64) *Node
	// Expand is used to create and select untried actions from the passed node.
	Expand(node *Node, explorationParam float64) *Node
}

// BackpropPolicy defines how a montcarlo tree search should propagate scores
// towards the root node.
type BackpropPolicy interface {
	Backpropagate(node *Node, score float64)
}

// Policy is an interface containing all sub-policies required to define a MCTS.
type Policy interface {
	DefaultPolicy
	TreePolicy
	BackpropPolicy
}
