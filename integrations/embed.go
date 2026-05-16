package integrations

import "embed"

// CursorRules holds installable Cursor user rules (alwaysApply).
//
//go:embed cursor/rules/*.mdc
var CursorRules embed.FS
