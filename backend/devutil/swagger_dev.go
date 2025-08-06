//go:build dev
// +build dev

package devutil

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	swgui "github.com/swaggest/swgui/v5emb"
)

// MountSwagger serves /openapi.yaml and swagger-ui at /docs/ for dev builds.
func MountSwagger(r chi.Router, title string) {
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})
	r.Mount("/docs/", swgui.New(title, "/openapi.yaml", "/docs/"))
}
