package montecarlo

import (
	"math/rand"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
)

// DeterminizationPolicy is a policy - an extension of the UCTPolicy - for the determinization of probabilistic
// actions - all nodes with this policy must have the ability to determine the
// probability of an action leading to a child node.
//
// DeterminizationPolicy should be used where a player has no control over which
// action is selected - and the only determining factor is a distinct probability
type DeterminizationPolicy struct {
	childProbability map[Key]float64
}

// NewDeterminizationPolicy constructs a determinization policy for determinizing
// the specified action keys, based on their probability
func NewDeterminizationPolicy(probabilities map[Key]float64) DeterminizationPolicy {
	dp := DeterminizationPolicy{}
	dp.childProbability = probabilities
	return dp
}

// probabilityPair is a pair of action key and its probability
type probabilityPair struct {
	probability float64
	action      Key
}

// probabilityPairList is simply an array of probabilityPairs
type probabilityPairList []probabilityPair

// ChildProbability returns the probability of a child node being selected
func (dp DeterminizationPolicy) ChildProbability(child Key) float64 {
	if p, ok := dp.childProbability[child]; ok {
		return p
	}
	return 0
}

/*-------- IMPLEMENT sort.Interface --------*/
// so that an array of probabilityPairs can be sorted according to probability
// https://groups.google.com/forum/#!topic/golang-nuts/FT7cjmcL7gw

func (ppl probabilityPairList) Swap(i, j int) {
	ppl[i], ppl[j] = ppl[j], ppl[i]
}

func (ppl probabilityPairList) Len() int {
	return len(ppl)
}

func (ppl probabilityPairList) Less(i, j int) bool {
	// compare by probability
	return ppl[i].probability < ppl[j].probability
}

/*-------- IMPLEMNET Policy --------*/

// Select the key for an action based on it's probability
func (dp DeterminizationPolicy) Select(node *Node, expl float64) *Node {
	if len(node.children) <= 0 && node.IsExhausted() {
		return nil
	}
	// generate all moves
	for !node.IsExhausted() {
		node.Policy().Expand(node, expl)
	}
	cProbability := float64(0)
	pairs := make(probabilityPairList, len(node.children))
	for k := range node.children {
		pairs = append(pairs, probabilityPair{
			probability: dp.ChildProbability(k),
			action:      k,
		})
	}
	// sort the list of pairs by their probability (lowest first)
	sort.Sort(pairs)
	// pick a random number as the target (in range [0, 1))
	rand.Seed(time.Now().UTC().UnixNano())
	target := rand.Float64()
	var actionKey *Key
	actionKey = nil
	for _, v := range pairs {
		cProbability += v.probability
		if float64(target) < cProbability {
			actionKey = &v.action
		}
	}
	if cProbability > 1 {
		log.Warnf("probabilities for children of deterministic node %v are cumulatively more probable than 1!", node)
	}
	return node.GetChild(actionKey)
}

// Simulate acts in exactly the same way as the UCTPolicy
func (dp DeterminizationPolicy) Simulate(node *Node) float64 {
	return UCTPolicy{}.Simulate(node)
}

// Expand acts in exactly the same way as the UCTPolicy
func (dp DeterminizationPolicy) Expand(node *Node, expl float64) *Node {
	return UCTPolicy{}.Expand(node, expl)
}

// Backpropagate acts in exactly the same way as the UCTPolicy
func (dp DeterminizationPolicy) Backpropagate(node *Node, score float64) {
	UCTPolicy{}.Backpropagate(node, score)
}
