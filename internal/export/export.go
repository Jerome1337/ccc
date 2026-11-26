// Package export copies the Claude Code config already on this machine into a folder laid out as a
// config repo, so `ccc --from` can install it elsewhere.
package export

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Jerome1337/ccc/internal/install"
	"github.com/Jerome1337/ccc/internal/term"
)

const (
	projectConfigDir = ".claude"
	fileMode         = 0o644
	dirMode          = 0o755
)

var (
	// Only config is copied: history, caches and settings.json (which can hold API keys) stay on this machine.
	configFolders = []string{"agents", "commands", "output-styles", "rules", "skills"}
	entryFiles    = map[install.Mode][]string{
		install.Global:  {"CLAUDE.md", "AGENTS.md", "ccc.yml"},
		install.Project: {"CLAUDE.md", "AGENTS.md"},
	}
	importPattern = regexp.MustCompile("(?:^|\\s)@([^\\s`]+)")
	// Claude Code syncs these skills from claude.ai on every machine you log in on.
	managedFolders = []string{"skills/synced"}
)

// Options copies the config of From (~/.claude or a project) into To/global or To/project.
type Options struct {
	From     string
	To       string
	Mode     install.Mode
	IsForced bool
	Out      io.Writer
}

// Run returns how many files it copied. It won't write into an existing folder unless IsForced.
func Run(options Options) (int, error) {
	dest := filepath.Join(options.To, install.SourceDir(options.Mode))
	if _, err := os.Stat(dest); err == nil && !options.IsForced {
		return 0, fmt.Errorf("%s already exists, remove it or pass --force to write over it", dest)
	}
	files, skipped, err := collect(options.From, options.Mode)
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 0, fmt.Errorf("found no Claude Code config in %s", options.From)
	}

	for _, file := range files {
		if err := copyFile(filepath.Join(options.From, file), filepath.Join(dest, file)); err != nil {
			return 0, err
		}
		fmt.Fprintf(options.Out, "  %s  %s\n", term.Green("copy  "), path.Join(install.SourceDir(options.Mode), file))
	}
	for _, name := range skipped {
		fmt.Fprintf(options.Out, "  %s  %s %s\n", term.Yellow("skip  "), name, term.Dim("(installed by another tool)"))
	}

	return len(files), nil
}

// collect lists the files to copy and the managed ones it leaves out, as slash paths relative to from.
func collect(from string, mode install.Mode) ([]string, []string, error) {
	files := withImports(from, entryFiles[mode])
	var skipped []string
	configDir := ""
	if mode == install.Project {
		configDir = projectConfigDir
	}
	for _, folder := range configFolders {
		nested, managed, err := walk(from, path.Join(configDir, folder))
		if err != nil {
			return nil, nil, err
		}
		files = append(files, nested...)
		skipped = append(skipped, managed...)
	}
	slices.Sort(files)

	return slices.Compact(files), skipped, nil
}

// withImports adds the files the given ones import with @path, as Claude Code reads them, and theirs in turn.
func withImports(from string, names []string) []string {
	var files []string
	seen := map[string]bool{}
	for len(names) > 0 {
		name := names[0]
		names = names[1:]
		if seen[name] || !filepath.IsLocal(filepath.FromSlash(name)) || !isFile(filepath.Join(from, name)) {
			continue
		}
		seen[name] = true
		files = append(files, name)
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(from, name))
		if err != nil {
			continue
		}
		for _, match := range importPattern.FindAllStringSubmatch(string(data), -1) {
			names = append(names, path.Join(path.Dir(name), strings.TrimRight(match[1], ".,;:!?)")))
		}
	}

	return files
}

// walk skips what another tool manages: a symlink in the folder (npx skills links its skills there)
// and the managed folders. A symlinked config folder, as in a dotfiles setup, is followed.
func walk(from, folder string) ([]string, []string, error) {
	root, err := filepath.EvalSymlinks(filepath.Join(from, folder))
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	var files, skipped []string
	err = filepath.WalkDir(root, func(current string, entry fs.DirEntry, err error) error {
		if err != nil || current == root {
			return err
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		name := path.Join(folder, filepath.ToSlash(relative))
		isManaged := entry.Type()&fs.ModeSymlink != 0 || slices.Contains(managedFolders, strings.TrimPrefix(name, projectConfigDir+"/"))
		if isManaged {
			skipped = append(skipped, name)
		}
		if isManaged || strings.HasPrefix(entry.Name(), ".") {
			return skipDir(entry)
		}
		if !entry.IsDir() {
			files = append(files, name)
		}

		return nil
	})

	return files, skipped, err
}

// skipDir leaves out a folder and everything in it, or just the one file.
func skipDir(entry fs.DirEntry) error {
	if entry.IsDir() {
		return filepath.SkipDir
	}

	return nil
}

func copyFile(source, dest string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), dirMode); err != nil {
		return err
	}

	return os.WriteFile(dest, data, fileMode)
}

func isFile(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.Mode().IsRegular()
}
