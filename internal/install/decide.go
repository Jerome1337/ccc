package install

import "github.com/Jerome1337/ccc/internal/lock"

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
	// ccc wrote the file and you edited it since. An entry with only a declined version means ccc never wrote it.
	if state.IsLocked && state.Locked.SHA256 != "" {
		return Release
	}

	return Forget
}

// isUnedited means the file still matches what the lock says ccc last wrote.
func isUnedited(state State) bool {
	return state.IsLocked && state.Locked.SHA256 != "" && state.Disk == state.Locked.SHA256
}
