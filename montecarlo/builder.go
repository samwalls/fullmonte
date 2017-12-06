package montecarlo

// ActionBuilder is a type that can sets of possible actions.
type ActionBuilder interface {
	BuildActions(defaultState State) ActionSet
}

// MasterBuilder builds all actions, based on a list of other ActionBuilders
type MasterBuilder struct {
	SubBuilders []ActionBuilder
}

// BuildActions builds all actions
func (mb MasterBuilder) BuildActions(defaultState State) ActionSet {
	actions := make(ActionSet)
	for _, builder := range mb.SubBuilders {
		actions.merge(builder.BuildActions(defaultState))
	}
	return actions
}

// add/overwrite all of the elements of b into this action builders current set
// of actions.
func (as ActionSet) merge(others ActionSet) {
	for k, v := range others {
		as[k] = v
	}
}
