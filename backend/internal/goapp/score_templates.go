package goapp

import (
	"net/http"
	"time"
)

const scoreTemplateConfigVersion = "2026-05-24"

func (a *App) handleGetScoreTemplates(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, success("score templates", map[string]any{
		"version":             scoreTemplateConfigVersion,
		"updated_at":          time.Now().UTC().Format(time.RFC3339),
		"resonator_templates": resonatorTemplates,
	}))
}
