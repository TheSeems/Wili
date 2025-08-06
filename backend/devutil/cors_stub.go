//go:build !dev
// +build !dev

package devutil

import "github.com/go-chi/chi/v5"

// EnableCORS is a no-op in non-dev builds.
func EnableCORS(_ chi.Router) {}
