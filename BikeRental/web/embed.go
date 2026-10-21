// Package web contains the embedded browser interface.
package web

import "embed"

//go:embed page.gohtml assets
var Files embed.FS
