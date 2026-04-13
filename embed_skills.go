// Package kankuro embeds canonical skill assets from the kankuro_skills directory
// so they are bundled into the binary. This file exists at the module root to allow
// embedding files that are siblings to the module root's subdirectories.
package kankuro

import (
	_ "embed"
)

// EmbeddedMCPCLISkill contains the embedded SKILL.md for the mcp-cli built-in skill.
// This embeds the canonical skill definition from kankuro_skills/mcp-cli/SKILL.md.
//
//go:embed kankuro_skills/mcp-cli/SKILL.md
var EmbeddedMCPCLISkill []byte
