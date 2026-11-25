package install

import "strings"

// legacyDests is where ccc 1.x (npm) installed files, so the first 2.x run can recognize them
// without a lock file. Versions 1.2.x also copied AGENTS.md into ~/.claude.
var legacyDests = map[Mode][]string{
	Project: {
		"CLAUDE.md", "AGENTS.md", "RTK.md", "CODING_STYLE.md",
		".claude/rules/architecture.md", ".claude/rules/code-style.md",
		".claude/rules/commands.md", ".claude/rules/patterns.md",
		".claude/skills/pr-review.md",
	},
	Global: {
		"CLAUDE.md", "AGENTS.md", "RTK.md", "CODING_STYLE.md", "ccc.yml",
		"rules/architecture.md", "rules/code-style.md", "rules/commands.md", "rules/patterns.md",
		"skills/pr-review.md",
	},
}

// legacyHashes is the sha256 of every version of each file shipped from v1.0.0 to v1.3.1,
// from `git show <tag>:<path> | shasum -a 256`.
var legacyHashes = map[string][]string{
	"AGENTS.md": {"488943731fbe0738490d7d73ff1d07f4683bc289ca02cd3b51c4c96fce5695ff"},
	"CLAUDE.md": {
		"11d544d95de0420418f13cb4e774e80e9697725d46576937877bb485d07809c5",
		"404e213143702ac8dedc655ea7ca8567f6b146eb47ebeb29449a5cb18062b172",
	},
	"CODING_STYLE.md":       {"89185323f59a92b07b896798b63eb87b3db084ee3452f0d948ba2ad6ed8b1cf1"},
	"RTK.md":                {"9fe007d9d7ca790f35b48bd12099e1b3a1dae05caf296dba3ddbe2621ddb02dc"},
	"ccc.yml":               {"e5de05def20c33b14eb085204083c9f9f21edf2387dfa188fb754dfbb2d68bef"},
	"rules/architecture.md": {"f774fce55b473e185cf488df63ba2420e94a0032bfb07bed8e7816f949d6ee2f"},
	"rules/code-style.md":   {"c6c7064e35a11c9abad1bc5d03f46a1ed1bb94c6af254841c652a06d9131fb2e"},
	"rules/commands.md":     {"e8d430e0e43bf27f31208a84a163e44a2343aef542535e4179730db27a72c088"},
	"rules/patterns.md":     {"d60cff709357e784edd06c10cd1dee3434751b23ef0db679ffa90458a368e581"},
	"skills/pr-review.md":   {"a915811e4284e07f16e45916e72342338f54deaae38d7492eb7880dc2a8a90b5"},
}

// legacyHashesFor maps an installed path back to the 1.x source file it was copied from.
func legacyHashesFor(dest string) []string {
	return legacyHashes[strings.TrimPrefix(dest, projectConfigDir+"/")]
}
