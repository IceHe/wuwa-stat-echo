package goapp

import (
	"net/http"
	"time"
)

const scoreTemplateConfigVersion = "xwuid-2026-09-14-compat-2026-10-06-v2"

func (a *App) handleGetScoreTemplates(w http.ResponseWriter, r *http.Request) {
	templates := make([]resonatorTemplate, 0, len(xwuidTemplates))
	for _, template := range xwuidTemplates {
		converted, ok := resonatorTemplates[template.Name]
		if ok {
			templates = append(templates, converted)
		}
	}
	writeJSON(w, success("score templates", map[string]any{
		"version":             scoreTemplateConfigVersion,
		"updated_at":          time.Now().UTC().Format(time.RFC3339),
		"templates":           templates,
		"resonator_templates": xwuidTemplateByKey,
	}))
}
