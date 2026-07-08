// Package web embeds the static UI assets so the binary is self-contained.
package web

import "embed"

// FS holds the embedded web assets (static/* and index.html).
//
//go:embed static index.html
var FS embed.FS
