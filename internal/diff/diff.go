// Package diff renders a small line diff, enough to review a config file before overwriting it.
package diff

import "strings"

type op int

const (
	equal op = iota
	deleted
	inserted
)

var prefixes = map[op]string{equal: "  ", deleted: "- ", inserted: "+ "}

type line struct {
	op   op
	text string
}

// Unified returns the changed lines of oldText → newText, prefixed with "- " and "+ ",
// with `context` unchanged lines around each change and "@@" between distant changes.
func Unified(oldText, newText string, context int) string {
	lines := compare(split(oldText), split(newText))
	isShown := markShown(lines, context)

	var out strings.Builder
	previous := -2
	for index, current := range lines {
		if !isShown[index] {
			continue
		}
		if index != previous+1 {
			out.WriteString("@@\n")
		}
		out.WriteString(prefixes[current.op] + current.text + "\n")
		previous = index
	}

	return out.String()
}

func split(text string) []string {
	if text == "" {
		return nil
	}

	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// compare walks a longest-common-subsequence table; config files are small, so O(n·m) is fine.
func compare(before, after []string) []line {
	common := commonLengths(before, after)
	lines := make([]line, 0, len(before)+len(after))
	i, j := 0, 0
	for i < len(before) || j < len(after) {
		switch {
		case i < len(before) && j < len(after) && before[i] == after[j]:
			lines = append(lines, line{equal, before[i]})
			i++
			j++
		case j == len(after) || (i < len(before) && common[i+1][j] >= common[i][j+1]):
			lines = append(lines, line{deleted, before[i]})
			i++
		default:
			lines = append(lines, line{inserted, after[j]})
			j++
		}
	}

	return lines
}

// commonLengths[i][j] is the length of the longest common subsequence of before[i:] and after[j:].
func commonLengths(before, after []string) [][]int {
	common := make([][]int, len(before)+1)
	for i := range common {
		common[i] = make([]int, len(after)+1)
	}
	for i := len(before) - 1; i >= 0; i-- {
		for j := len(after) - 1; j >= 0; j-- {
			if before[i] == after[j] {
				common[i][j] = common[i+1][j+1] + 1
				continue
			}
			common[i][j] = max(common[i+1][j], common[i][j+1])
		}
	}

	return common
}

func markShown(lines []line, context int) []bool {
	isShown := make([]bool, len(lines))
	for index, current := range lines {
		if current.op == equal {
			continue
		}
		for near := max(0, index-context); near <= min(len(lines)-1, index+context); near++ {
			isShown[near] = true
		}
	}

	return isShown
}
