package montecarlo

// Tree contains all the information needed to progress a MCTS: a root
// montecarlo.Node and a set of possible actions.
type Tree struct {
	possibleActions map[Key]Action
	root            Node
}

// NewTree is a constructor for a montecarlo.Tree struct.
func NewTree(numPlayers uint, initialState State, possibleActions map[Key]Action) (Tree, error) {
	node, err := NewNode(numPlayers)
	node.State = initialState
	return Tree{
		root:            node,
		possibleActions: possibleActions,
	}, err
}

// Root returns the root montecarlo.Node of the tree.
func (tree *Tree) Root() *Node {
	return &tree.root
}

// PossibleActions returns the set of possible actions defined in the
// montecarlo.Tree struct.
func (tree *Tree) PossibleActions() map[Key]Action {
	return tree.possibleActions
}

// Copy creates a deep copy of this tree.
func (tree *Tree) Copy() *Tree {
	actions := make(map[Key]Action)
	for k, v := range tree.possibleActions {
		actions[k] = v
	}
	root := tree.Root()
	rootCpy := root.Copy()
	// will not throw any error since we're already using a valid player count
	cpy, _ := NewTree(root.NumPlayers(), rootCpy.State, actions)
	cpy.root = *rootCpy
	return &cpy
}

// Merge two trees together: add all nodes from other into this tree. If both
// trees have the same node, then their Score and Visit values are added.
func (tree *Tree) Merge(other *Tree) error {
	root := tree.Root()
	return root.Merge(other.Root())
}
