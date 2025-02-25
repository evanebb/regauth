package handlers

import (
	"github.com/evanebb/regauth/template"
	"net/http"
)

func ManageAccount(t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.RenderBase(w, r, nil, "account/manage.gohtml")
	}
}
