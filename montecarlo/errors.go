package montecarlo

import (
	"fmt"
)

/*
 All error types for the montecarlo library. The following types contain
 information required to report the error.
*/

// ZeroPlayerCount thrown when a Tree or Node is instantiated with an 0 players.
type ZeroPlayerCount Node

// MergeDifferingPlayerCount thrown when a tree is merged with another that has
// a differing player count.
type MergeDifferingPlayerCount struct {
	one   uint
	other uint
}

// MergePlayerIndexMismatch thrown when two tree nodes by comparison have a differing
// account of their player index, when they were expected to be the same
type MergePlayerIndexMismatch struct {
	one   uint
	other uint
}

// MergeStateMismatch thrown when merging two trees and states which follow the
// same path differ.
type MergeStateMismatch struct {
	one   State
	other State
}

/*
 Implement the Error interface for all the error types.
*/

func (zpc ZeroPlayerCount) Error() string {
	return fmt.Sprintf("can't create node with zero players: %v", zpc)
}

func (mpdc MergeDifferingPlayerCount) Error() string {
	return fmt.Sprintf("can't merge tree of player count %v with a tree of player count %v", mpdc.one, mpdc.other)
}

func (mpim MergePlayerIndexMismatch) Error() string {
	return fmt.Sprintf("player index mismatch: %v vs. %v", mpim.one, mpim.other)
}

func (msm MergeStateMismatch) Error() string {
	return fmt.Sprintf("merge state mismatch: %v vs. %v", msm.one, msm.other)
}
