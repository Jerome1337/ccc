package install

import (
	"testing"

	"github.com/Jerome1337/ccc/internal/lock"
)

func TestDecide(t *testing.T) {
	const shipped, mine, previous, legacy = "shipped", "mine", "previous", "legacy"
	update := Context{IsUpdate: true, CanPrompt: true}
	cases := []struct {
		name    string
		state   State
		context Context
		want    Action
	}{
		{"missing file", State{IsShipped: true, Shipped: shipped}, update, Create},
		{"up to date", State{IsShipped: true, Shipped: shipped, Disk: shipped}, update, Keep},
		{"unedited by lock", State{IsShipped: true, Shipped: shipped, Disk: previous, IsLocked: true, Locked: lock.File{SHA256: previous}}, Context{}, Overwrite},
		{"unedited by 1.x hash, no lock entry", State{IsShipped: true, Shipped: shipped, Disk: legacy, Legacy: []string{legacy}}, Context{}, Overwrite},
		{"1.x hash ignored once the lock tracks the file", State{IsShipped: true, Shipped: shipped, Disk: legacy, IsLocked: true, Locked: lock.File{SHA256: previous}, Legacy: []string{legacy}}, update, Ask},
		{"edited, update in a terminal", State{IsShipped: true, Shipped: shipped, Disk: mine, IsLocked: true, Locked: lock.File{SHA256: previous}}, update, Ask},
		{"edited, update without a terminal", State{IsShipped: true, Shipped: shipped, Disk: mine}, Context{IsUpdate: true}, Skip},
		{"edited, init in a terminal", State{IsShipped: true, Shipped: shipped, Disk: mine}, Context{CanPrompt: true}, Skip},
		{"edited, forced", State{IsShipped: true, Shipped: shipped, Disk: mine}, Context{IsForced: true}, Overwrite},
		{"edited, already declined this version", State{IsShipped: true, Shipped: shipped, Disk: mine, IsLocked: true, Locked: lock.File{SHA256: previous, Declined: shipped}}, update, Declined},
		{"edited, declined an older version", State{IsShipped: true, Shipped: shipped, Disk: mine, IsLocked: true, Locked: lock.File{Declined: "older"}}, update, Ask},
		{"forced beats declined", State{IsShipped: true, Shipped: shipped, Disk: mine, IsLocked: true, Locked: lock.File{Declined: shipped}}, Context{IsForced: true}, Overwrite},
		{"declined-only entry isn't proof of ccc's version", State{IsShipped: true, Shipped: shipped, Disk: mine, IsLocked: true, Locked: lock.File{Declined: "older"}}, Context{}, Skip},
		{"no longer shipped, unedited", State{Disk: previous, IsLocked: true, Locked: lock.File{SHA256: previous}}, update, Remove},
		{"no longer shipped, 1.x file", State{Disk: legacy, Legacy: []string{legacy}}, update, Remove},
		{"no longer shipped, edited", State{Disk: mine, IsLocked: true, Locked: lock.File{SHA256: previous}}, update, Release},
		{"no longer shipped, unknown file at a 1.x path", State{Disk: mine, Legacy: []string{legacy}}, update, Forget},
		{"no longer shipped, already gone", State{IsLocked: true, Locked: lock.File{SHA256: previous}}, update, Forget},
		{"no longer shipped, forced still keeps edits", State{Disk: mine, IsLocked: true, Locked: lock.File{SHA256: previous}}, Context{IsForced: true}, Release},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := Decide(testCase.state, testCase.context); got != testCase.want {
				t.Errorf("Decide() = %v, want %v", got, testCase.want)
			}
		})
	}
}
