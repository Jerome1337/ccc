package install

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Jerome1337/ccc/internal/diff"
	"github.com/Jerome1337/ccc/internal/lock"
	"github.com/Jerome1337/ccc/internal/term"
)

const (
	diffContext = 3
	// diffIndent lines the diff and the question up under the file name.
	diffIndent = "          "
	fileMode   = 0o644
	dirMode    = 0o755
)

type Prompter interface {
	CanAsk() bool
	Confirm(question string) bool
}

type Options struct {
	Content fs.FS
	// Source is the config repo Content was read from, saved in the lock for the next run.
	Source   string
	Target   string
	Mode     Mode
	IsUpdate bool
	IsForced bool
	Prompter Prompter
	Out      io.Writer
	Version  string
}

// Result counts files ccc wrote or deleted, and edited files it left alone.
type Result struct {
	Changed int
	Skipped int
}

type syncer struct {
	options Options
	lock    lock.Lock
	result  Result
}

type change struct {
	dest    string
	shipped []byte
	disk    []byte
	state   State
}

// Sync brings the target folder to the shipped version and saves the lock, even after a partial failure.
func Sync(options Options) (Result, error) {
	entries, err := Manifest(options.Content, options.Mode)
	if err != nil {
		return Result{}, err
	}
	path := lockPath(options.Target, options.Mode)
	current, exists, err := lock.Read(path)
	if err != nil {
		return Result{}, err
	}

	if exists && current.Source != options.Source {
		fmt.Fprintf(options.Out, "  %s  %s → %s\n\n", term.Cyan("switch"), sourceName(current.Source), sourceName(options.Source))
	}
	syncer := &syncer{options: options, lock: current}
	err = syncer.run(entries, syncer.unshippedPaths(entries))
	syncer.lock.Version = options.Version
	syncer.lock.Source = options.Source

	return syncer.result, errors.Join(err, lock.Write(path, syncer.lock))
}

func sourceName(source string) string {
	if source == "" {
		return "the built-in config"
	}

	return source
}

func (syncer *syncer) run(entries []Entry, unshipped []string) error {
	for _, entry := range entries {
		if err := syncer.syncShipped(entry); err != nil {
			return err
		}
	}
	for _, dest := range unshipped {
		if err := syncer.syncUnshipped(dest); err != nil {
			return err
		}
	}

	return nil
}

// unshippedPaths is what ccc tracked before but no longer ships.
// Paths that would leave the target folder are dropped, so a hand-edited lock can't delete files elsewhere.
func (syncer *syncer) unshippedPaths(entries []Entry) []string {
	candidates := slices.Collect(maps.Keys(syncer.lock.Files))

	isShippedOrOutside := func(dest string) bool {
		isShipped := slices.ContainsFunc(entries, func(entry Entry) bool { return entry.Dest == dest })

		return isShipped || !filepath.IsLocal(filepath.FromSlash(dest))
	}
	candidates = slices.DeleteFunc(candidates, isShippedOrOutside)
	slices.Sort(candidates)

	return slices.Compact(candidates)
}

func (syncer *syncer) syncShipped(entry Entry) error {
	shipped, err := fs.ReadFile(syncer.options.Content, entry.Source)
	if err != nil {
		return err
	}
	disk, diskHash, err := readDisk(toOSPath(syncer.options.Target, entry.Dest))
	if err != nil {
		return err
	}

	locked, isLocked := syncer.lock.Files[entry.Dest]
	state := State{
		IsShipped: true,
		Shipped:   hash(shipped),
		Disk:      diskHash,
		Locked:    locked,
		IsLocked:  isLocked,
	}
	current := change{dest: entry.Dest, shipped: shipped, disk: disk, state: state}

	return syncer.applyShipped(current, Decide(state, syncer.context()))
}

func (syncer *syncer) applyShipped(current change, action Action) error {
	switch action {
	case Keep:
		syncer.track(current.dest, current.state.Shipped)
		syncer.print(term.Dim("ok    "), current.dest, "")
	case Create:
		return syncer.write(current, term.Green("copy  "))
	case Overwrite:
		return syncer.write(current, term.Green("update"))
	case Ask:
		return syncer.ask(current)
	case Declined:
		syncer.skip(current.dest, "edited, keeping yours")
	case Skip:
		syncer.skip(current.dest, syncer.skipReason())
	}

	return nil
}

func (syncer *syncer) syncUnshipped(dest string) error {
	path := toOSPath(syncer.options.Target, dest)
	_, diskHash, err := readDisk(path)
	if err != nil {
		return err
	}

	locked, isLocked := syncer.lock.Files[dest]
	state := State{Disk: diskHash, Locked: locked, IsLocked: isLocked}
	delete(syncer.lock.Files, dest)

	switch Decide(state, syncer.context()) {
	case Remove:
		if err := os.Remove(path); err != nil {
			return err
		}
		removeEmptyParents(filepath.Dir(path), configRoot(syncer.options.Target, syncer.options.Mode))
		syncer.print(term.Cyan("delete"), dest, "no longer shipped")
		syncer.result.Changed++
	case Release:
		syncer.print(term.Yellow("keep  "), dest, "edited, no longer managed by ccc")
		syncer.result.Skipped++
	}

	return nil
}

func (syncer *syncer) ask(current change) error {
	syncer.print(term.Yellow("diff  "), current.dest, "- yours, + new version")
	changes := strings.TrimSuffix(diff.Unified(string(current.disk), string(current.shipped), diffContext), "\n")
	for _, line := range strings.Split(changes, "\n") {
		fmt.Fprintln(syncer.options.Out, diffIndent+colorDiffLine(line))
	}

	if syncer.options.Prompter.Confirm(fmt.Sprintf("%soverwrite %s?", diffIndent, current.dest)) {
		return syncer.write(current, term.Green("update"))
	}
	entry := syncer.lock.Files[current.dest]
	entry.Declined = current.state.Shipped
	syncer.lock.Files[current.dest] = entry
	syncer.skip(current.dest, "keeping yours")

	return nil
}

func colorDiffLine(line string) string {
	switch {
	case strings.HasPrefix(line, "- "):
		return term.Red(line)
	case strings.HasPrefix(line, "+ "):
		return term.Green(line)
	case strings.HasPrefix(line, "@@"):
		return term.Cyan(line)
	}

	return line
}

func (syncer *syncer) write(current change, label string) error {
	path := toOSPath(syncer.options.Target, current.dest)
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return err
	}
	if err := os.WriteFile(path, current.shipped, fileMode); err != nil {
		return err
	}
	syncer.track(current.dest, current.state.Shipped)
	syncer.print(label, current.dest, "")
	syncer.result.Changed++

	return nil
}

func (syncer *syncer) track(dest, sha string) {
	syncer.lock.Files[dest] = lock.File{SHA256: sha}
}

func (syncer *syncer) skip(dest, reason string) {
	syncer.print(term.Yellow("skip  "), dest, reason)
	syncer.result.Skipped++
}

func (syncer *syncer) skipReason() string {
	if syncer.options.IsUpdate {
		return "edited, run ccc update in a terminal to review it"
	}

	return "edited, run ccc update to review it, or --force to overwrite"
}

func (syncer *syncer) context() Context {
	return Context{
		IsUpdate:  syncer.options.IsUpdate,
		CanPrompt: syncer.options.Prompter.CanAsk(),
		IsForced:  syncer.options.IsForced,
	}
}

func (syncer *syncer) print(label, dest, detail string) {
	if detail == "" {
		fmt.Fprintf(syncer.options.Out, "  %s  %s\n", label, dest)

		return
	}
	fmt.Fprintf(syncer.options.Out, "  %s  %s %s\n", label, dest, term.Dim("("+detail+")"))
}

// readDisk returns the file and its hash; the hash is empty when the file doesn't exist.
func readDisk(path string) ([]byte, string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	return data, hash(data), nil
}

func hash(data []byte) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

// removeEmptyParents deletes folders left empty by a removal, never the config root itself.
func removeEmptyParents(dir, root string) {
	for dir != root && strings.HasPrefix(dir, root+string(filepath.Separator)) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}
