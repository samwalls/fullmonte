package montecarlo

import (
	"fmt"
	"math"
	"testing"

	log "github.com/Sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

/*-------- TEST INPUTS & SETUP --------*/

var normal, nodeWithChildren, nodeWithGrandchildren Node

func nodeTestSetup() {
	var err error
	normal, err = NewNode(1)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	nodeWithChildren, err = NewNode(1)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	nodeWithChildren.visits = 10
	nodeWithGrandchildren, err = NewNode(1)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	nodeWithGrandchildren.visits = 100
	for i := 0; i < 10; i++ {
		childNode, err := NewNode(1)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
		//add scores to create some sort of identifiable order
		childNode.SetScore(0, float64(i))
		childNode.visits = 1
		nodeWithChildren.SetChild(fmt.Sprintf("%v", i), &childNode)
		//add grandchildren and their respective parents
		parentNode, err := NewNode(1)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
		parentNode.SetScore(0, float64(i))
		parentNode.visits = 10
		for j := 0; j < 10; j++ {
			//add "grandchildren" to each child
			grandchild, err := NewNode(1)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
			grandchild.SetScore(0, float64(i*j))
			grandchild.visits = 1
			parentNode.SetChild(fmt.Sprintf("%v", j), &grandchild)
		}
		nodeWithGrandchildren.SetChild(fmt.Sprintf("%v", i), &parentNode)
	}
	log.SetLevel(log.DebugLevel)
}

/*-------- TESTING --------*/

func TestNewNode(t *testing.T) {
	nodeTestSetup()
	assert.Equal(t, float64(0), normal.Score(0), "score of new mcts node should be zero")
	assert.Equal(t, int64(0), normal.Visits(), "total sims of new mcts node should be zero")
	assert.Nil(t, normal.parent, "a new bare mcts node should have nil parent")
	assert.Equal(t, 0, len(normal.children), "new mcts node should have no children")
	assert.False(t, normal.Visits() > 0, "new mcts node should not be labelled visited")
	assert.Equal(t, 10, len(nodeWithChildren.children))
	assert.Equal(t, 10, len(nodeWithGrandchildren.children))
	for _, c := range nodeWithGrandchildren.children {
		assert.Equal(t, 10, len(c.children))
	}
}

func TestNewNodeIsRoot(t *testing.T) {
	nodeTestSetup()
	assert.True(t, normal.IsRoot(), "a new bare mcts node should have nil parent, and hence be a root")
}

func TestMakeZeroPlayerNode(t *testing.T) {
	_, err := NewNode(0)
	_, ok := err.(ZeroPlayerCount)
	if !ok {
		assert.Fail(t, "expected ZeroPlayerCount error when making node with zero players")
	}
	//try where there shouldn't be an error
	_, err = NewNode(1)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("%v", err))
	}
}

func TestSetAndGetScore(t *testing.T) {
	n, err := NewNode(1)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("%v", err))
	}
	assert.Equal(t, float64(0), n.Score(0))
	n.SetScore(0, 42)
	assert.Equal(t, float64(42), n.Score(0))
}

func TestSetAndGetScoreMultiplayer(t *testing.T) {
	n, err := NewNode(4)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("%v", err))
	}
	assert.Equal(t, []float64{0, 0, 0, 0}, n.ScoreVector())
	assert.Equal(t, float64(0), n.Score(0))
	assert.Equal(t, float64(0), n.Score(1))
	assert.Equal(t, float64(0), n.Score(2))
	assert.Equal(t, float64(0), n.Score(3))
	n.SetScore(0, 5)
	n.SetScore(1, 7)
	n.SetScore(2, 6)
	n.SetScore(3, 8)
	assert.Equal(t, []float64{5, 7, 6, 8}, n.ScoreVector())
	assert.Equal(t, float64(5), n.Score(0))
	assert.Equal(t, float64(7), n.Score(1))
	assert.Equal(t, float64(6), n.Score(2))
	assert.Equal(t, float64(8), n.Score(3))
}

func TestAddChild(t *testing.T) {
	nodeTestSetup()
	//try our test inputs
	prev := len(normal.children)
	newNode, err := NewNode(1)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("%v", err))
	}
	normal.SetChild("1", &newNode)
	assert.Equal(t, prev+1, len(normal.children), "adding child should increase NumChildren by 1")
	assert.NotNil(t, normal.GetChild("1"))
	assert.Equal(t, &normal, newNode.Parent(), "adding a child should set its parent")

	prev = len(nodeWithChildren.children)
	newNode, err = NewNode(1)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("%v", err))
	}
	nodeWithChildren.SetChild("11", &newNode)
	assert.Equal(t, prev+1, len(nodeWithChildren.children), "adding child should increase NumChildren by 1")
	assert.NotNil(t, nodeWithChildren.GetChild("11"))

	prev = len(nodeWithGrandchildren.children)
	newNode, err = NewNode(1)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("%v", err))
	}
	nodeWithGrandchildren.SetChild("11", &newNode)
	assert.Equal(t, prev+1, len(nodeWithGrandchildren.children), "adding child should increase NumChildren by 1")
	assert.NotNil(t, nodeWithGrandchildren.GetChild("11"))
}

func TestChildOrder(t *testing.T) {
	nodeTestSetup()
	assert.Equal(t, 10, len(nodeWithChildren.children), "mcts node should have indicated number of children")
	for i := 0; i < len(nodeWithChildren.children); i++ {
		//scores of children should be 1 to the number of children
		assert.Equal(t, float64(i), nodeWithChildren.GetChild(fmt.Sprintf("%v", i)).Score(0), "order of insertion should be preserved in mcts node")
		for j := 0; j < len(nodeWithChildren.children); j++ {
			assert.Equal(t, float64(i*j), nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).GetChild(fmt.Sprintf("%v", j)).Score(0), "order of insertion should be preserved in mcts node")
		}
	}
}

func TestRemoveChild(t *testing.T) {
	nodeTestSetup()
	copy := normal
	normal.RemoveChild("0")
	assert.Equal(t, copy, normal, "removing child from mcts node with no children should not affect it")

	prev := len(nodeWithChildren.children)
	child := nodeWithChildren.GetChild("0")
	assert.NotNil(t, child, "node should have returned valid child")
	nodeWithChildren.RemoveChild("0")
	assert.Equal(t, prev-1, len(nodeWithChildren.children), "removing from mcts node should decrease number of children")
	assert.Nil(t, child.Parent(), "removing child should cause child's parent to be nil")
	assert.Nil(t, nodeWithChildren.GetChild("0"), "child was removed, hence GetChild should return nil")
}

func TestGetChild(t *testing.T) {
	nodeTestSetup()
	assert.Nil(t, normal.GetChild("0"), "mcts GetChild should return nil if there are no children")
	assert.Nil(t, normal.GetChild("0"), "mcts GetChild should return nil if there are no children")
	assert.Nil(t, normal.GetChild("2"), "mcts GetChild should return nil if there are no children")
	assert.Nil(t, normal.GetChild("-1"), "mcts GetChild should return nil if there are no children")
}

func TestNumChildren(t *testing.T) {
	nodeTestSetup()
	assert.Equal(t, 0, len(normal.children), "new mcts node should have no children")
	assert.Equal(t, 10, len(nodeWithChildren.children), "mcts node should have indicated number of children")
	assert.Equal(t, 10, len(nodeWithGrandchildren.children), "mcts node should have indicated number of children")
	//test children too
	for i := 0; i < len(nodeWithGrandchildren.children); i++ {
		assert.Equal(t, 10, len(nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).children), "mcts node should have indicated number of children")
	}
}

func TestIsTerminal(t *testing.T) {
	nodeTestSetup()
	assert.True(t, normal.IsLeaf(), "bare mcts node should be terminal")
	assert.False(t, nodeWithChildren.IsLeaf(), "mcts nodes with children should not be terminal")
	assert.False(t, nodeWithGrandchildren.IsLeaf(), "mcts nodes with children should not be terminal")
	for i := 0; i < len(nodeWithChildren.children); i++ {
		assert.True(t, nodeWithChildren.GetChild(fmt.Sprintf("%v", i)).IsLeaf(), "mcts nodes without children should be terminal")
		assert.False(t, nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).IsLeaf(), "mcts nodes with children should not be terminal")
		for j := 0; j < len(nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).children); j++ {
			assert.True(t, nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).GetChild(fmt.Sprintf("%v", j)).IsLeaf(), "mcts nodes without children should be terminal")
		}
	}
}

func TestParent(t *testing.T) {
	nodeTestSetup()
	assert.Nil(t, normal.Parent(), "a bare mcts node should have no parent")
	for i := 0; i < len(nodeWithChildren.children); i++ {
		assert.Equal(t, &nodeWithChildren, nodeWithChildren.GetChild(fmt.Sprintf("%v", i)).Parent(), "a child mcts node should know its parent")
		assert.Equal(t, &nodeWithGrandchildren, nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).Parent(), "a child mcts node should know its parent")
		for j := 0; j < len(nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).children); j++ {
			assert.Equal(t, nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)), nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).GetChild(fmt.Sprintf("%v", j)).Parent(), "a grandchild mcts node should know its parent")
			assert.Equal(t, &nodeWithGrandchildren, nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", i)).GetChild(fmt.Sprintf("%v", j)).Parent().Parent(), "a grandchild mcts node should know its grandparent, via its parent")
		}
	}
}

func TestNodeSelectBestChild(t *testing.T) {
	nodeTestSetup()
	// TODO currently based on pure exploitation (no exploration)
	_, node := nodeWithGrandchildren.selectBestChild(0)
	assert.Equal(t, float64(9), node.Score(0))
	_, node = nodeWithGrandchildren.selectBestChild(0)
	assert.Equal(t, float64(9), node.Score(0))
}

func TestSelectBestChildDirectFromRoot(t *testing.T) {
	nodeTestSetup()
	k, node := nodeWithChildren.selectBestChild(0)
	assert.NotEqual(t, "", k)
	assert.NotNil(t, node)
}

func TestSelectBestChildAsLeaf(t *testing.T) {
	nodeTestSetup()
	grandChild := nodeWithGrandchildren.GetChild(fmt.Sprintf("%v", 0)).GetChild(fmt.Sprintf("%v", 0))
	k, c := grandChild.selectBestChild(0)
	assert.Equal(t, "", k)
	assert.Equal(t, grandChild, c)
}

func TestUCBExplorationParamLessThanZero(t *testing.T) {
	nodeTestSetup()
	ucb := nodeWithGrandchildren.UpperConfidenceBound(float64(-1), 0)
	assert.Equal(t, math.Inf(1), ucb)
}

func TestUpperConfidenceBoundDirectFromRoot(t *testing.T) {
	nodeTestSetup()
	ucb := normal.UpperConfidenceBound(0, 0)
	assert.Equal(t, math.Inf(1), ucb)
	ucb = nodeWithChildren.UpperConfidenceBound(0, 0)
	assert.Equal(t, math.Inf(1), ucb)
	for _, c := range nodeWithChildren.children {
		ucb = c.UpperConfidenceBound(0, 0)
		assert.Equal(t, math.Inf(1), ucb)
	}
}

func TestNodeIsExhausted(t *testing.T) {
	nodeTestSetup()
	assert.True(t, normal.IsExhausted())
	assert.True(t, nodeWithChildren.IsExhausted())
	assert.True(t, nodeWithGrandchildren.IsExhausted())
}

func TestNodeString(t *testing.T) {
	nodeTestSetup()
	assert.NotEqual(t, "", nodeWithGrandchildren.String(), "string output is empty")
}

func TestNodeAddVisit(t *testing.T) {
	node, err := NewNode(1)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	v := node.Visits()
	node.AddVisit()
	assert.Equal(t, v+1, node.Visits())
}

func TestNodeCopy(t *testing.T) {
	nodeTestSetup()
	cpy := nodeWithGrandchildren.Copy()
	assert.NotEqual(t, nodeWithGrandchildren, cpy)
	assert.Equal(t, nodeWithGrandchildren.State, cpy.State)
	for k, c := range nodeWithGrandchildren.children {
		cpyChild, ok := cpy.children[k]
		assert.True(t, ok)
		assert.NotEqual(t, &c, cpyChild)
		assert.Equal(t, c.State, cpyChild.State)
		for k2, c2 := range c.children {
			cpyChild, ok := cpy.children[k].children[k2]
			assert.True(t, ok)
			assert.NotEqual(t, &c2, cpyChild)
			assert.Equal(t, c2.State, cpyChild.State)
		}
	}
}

func TestNodeMergeWithSelf(t *testing.T) {
	nodeTestSetup()
	cpy := nodeWithGrandchildren.Copy()
	err := nodeWithGrandchildren.Merge(*cpy)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	//expect visits and score to be doubled
	v := cpy.Visits()
	assert.Equal(t, v*2, nodeWithGrandchildren.Visits())
	s := cpy.ScoreVector()
	for i, v := range nodeWithGrandchildren.ScoreVector() {
		assert.Equal(t, v*2, s[i])
	}
	assert.Equal(t, len(cpy.children), len(nodeWithGrandchildren.children))
}

func TestNodeMergeDifferingPlayerCount(t *testing.T) {
	nodeTestSetup()
	node, err := NewNode(5)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	err = nodeWithGrandchildren.Merge(node)
	assert.NotNil(t, err, "expected MergeDifferingPlayercount error when merging")
	var ok bool
	err, ok = err.(MergeDifferingPlayerCount)
	assert.True(t, ok, "expected MergeDifferingPlayercount error when merging")
}

func TestNodeMergePlayerIndexMismatch(t *testing.T) {
	nodeTestSetup()
	nodeWithGrandchildren.State = simpleStateImplementation{
		triggerPlayerIndexError: false,
	}
	node, err := NewNode(1)
	node.State = simpleStateImplementation{
		triggerPlayerIndexError: true,
	}
	if err != nil {
		assert.Fail(t, err.Error())
	}
	err = nodeWithGrandchildren.Merge(node)
	assert.NotNil(t, err, "expected error when merging")
	var ok bool
	err, ok = err.(MergePlayerIndexMismatch)
	assert.True(t, ok, "expected error to be of type MergePlayerIndexMismatch")
}

// very simple state implementation to test with all nodes using this state will
// be exhausted and terminal
type simpleStateImplementation struct {
	internal                int
	triggerPlayerIndexError bool
}

func (ssi simpleStateImplementation) LegalActions() ActionSet {
	return make(ActionSet)
}

func (ssi simpleStateImplementation) Score(player uint) float64 {
	return float64(0)
}

func (ssi simpleStateImplementation) Bias() float64 {
	return float64(0)
}

func (ssi simpleStateImplementation) Copy() State {
	return ssi
}

func (ssi simpleStateImplementation) Player() uint {
	if ssi.triggerPlayerIndexError {
		return 1
	}
	return 0
}

func (ssi simpleStateImplementation) Policy() Policy {
	return nil
}

func TestNodeIsTerminal(t *testing.T) {
	nodeTestSetup()
	assert.True(t, normal.IsTerminal())
}

func TestNodeWithStateIsExhausted(t *testing.T) {
	nodeTestSetup()
	normal.State = simpleStateImplementation{}
	assert.True(t, normal.IsExhausted())
	nodeWithChildren.State = simpleStateImplementation{}
	assert.True(t, nodeWithChildren.IsExhausted())
	nodeWithGrandchildren.State = simpleStateImplementation{}
	assert.True(t, nodeWithGrandchildren.IsExhausted())
}
