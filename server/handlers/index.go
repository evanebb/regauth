package handlers

import (
	"github.com/evanebb/regauth/template"
	"net/http"
)

func Index(t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.RenderBase(w, r, nil, "home.gohtml")
	}
}
