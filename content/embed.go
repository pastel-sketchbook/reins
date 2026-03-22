// Package content embeds the managed, scaffold, and template files that
// the reins CLI ships to downstream projects.
package content

import "embed"

// FS contains all embedded content.
//
// Directory layout:
//
//	managed/    — files owned by reins, refreshed via `reins update`
//	scaffold/   — project files auto-copied during `reins init` (skip if exists)
//	templates/  — language templates available for manual copying
//
//go:embed all:managed all:scaffold all:templates
var FS embed.FS
