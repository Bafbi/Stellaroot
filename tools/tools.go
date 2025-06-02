//go:build tools

package tools

// These imports only exist to keep go.mod entries for packages that are referenced in BUILD files,
// but not in Go code.

import (
	_ "github.com/a-h/templ/cmd/templ"
)
