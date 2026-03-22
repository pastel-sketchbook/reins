// Package content embeds the managed, scaffold, template, and skill files
// that the reins CLI ships to downstream projects.
package content

import "embed"

// FS contains all embedded content.
//
// Directory layout:
//
//	managed/    — files owned by reins, refreshed via `reins update`
//	scaffold/   — project files auto-copied during `reins init` (skip if exists)
//	templates/  — language templates available for manual copying
//	skill/      — AI tool skill definition installed globally or locally
//
//go:embed all:managed all:scaffold all:templates all:skill
var FS embed.FS
