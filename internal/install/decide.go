package install

import (
	"slices"

	"github.com/Jerome1337/ccc/internal/lock"
)

type Action int

const (
	// Keep: the file already matches the shipped version.
	Keep Action = iota
	// Create: the file is missing.
	Create
	// Overwrite: the file is ccc's own (unedited), or --force was passed.
	Overwrite
	// Ask: you edited the file and a terminal can show the diff and ask.
	Ask
	// Skip: you edited the file and nobody can be asked (no terminal, or init).
	Skip
	// Declined: you already said no to this exact shipped version.
	Declined
	// Remove: ccc no longer ships the file and you never edited it.
	Remove
	// Release: ccc no longer ships the file, but you edited it, so it stays and is no longer managed.
	Release
	// Forget: ccc no longer ships the file and has nothing to do: it's gone, or ccc never wrote it.
	Forget
)

// State describes one path. Hashes are sha256 hex; Disk is empty when the file doesn't exist.
type State struct {
	IsShipped bool
	Shipped   string
	Disk      string
	Locked    lock.File
	IsLocked  bool
	Legacy    []string
}

type Context struct {
	IsUpdate  bool
	CanPrompt bool
	IsForced  bool
}

// Decide is pure on purpose: every rule of the sync is one table test away.
func Decide(state State, context Context) Action {
	if !state.IsShipped {
		return decideUnshipped(state)
	}
	if state.Disk == "" {
		return Create
	}
	if state.Disk == state.Shipped {
		return Keep
	}
	if isUnedited(state) || context.IsForced {
		return Overwrite
	}
	if state.Locked.Declined == state.Shipped {
		return Declined
	}
	if context.IsUpdate && context.CanPrompt {
		return Ask
	}

	return Skip
}

func decideUnshipped(state State) Action {
	if state.Disk == "" {
		return Forget
	}
	if isUnedited(state) {
		return Remove
	}
	// Without a lock entry, a file at an old 1.x path that matches no 1.x hash may not be ccc's at all.
	if state.IsLocked {
		return Release
	}

	return Forget
}

// isUnedited trusts the lock when there is one, and the 1.x hashes only for files ccc has never tracked.
func isUnedited(state State) bool {
	if state.IsLocked {
		return state.Locked.SHA256 != "" && state.Disk == state.Locked.SHA256
	}

	return slices.Contains(state.Legacy, state.Disk)
}
