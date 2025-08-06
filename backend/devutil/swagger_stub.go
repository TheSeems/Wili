//go:build !dev
// +build !dev

package devutil

import "github.com/go-chi/chi/v5"

// MountSwagger is a no-op in prod builds.
func MountSwagger(_ chi.Router, _ string) {}
