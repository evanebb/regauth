package template

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/user"
	"github.com/gorilla/sessions"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
)

type Templater struct {
	logger       *slog.Logger
	templates    fs.FS
	sessionStore sessions.Store
}

func NewTemplater(logger *slog.Logger, templates fs.FS, sessionStore sessions.Store) Templater {
	return Templater{logger: logger, templates: templates, sessionStore: sessionStore}
}

func (t Templater) renderErr(w http.ResponseWriter, r *http.Request, data any, templateName string, files ...string) error {
	var err error

	session, _ := t.sessionStore.Get(r, "session")
	flashes := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	var uPtr *user.User
	u, err := getUserFromRequestContext(r.Context())
	if err == nil {
		uPtr = &u
	}

	currentUrl := r.URL.Path

	funcs := template.FuncMap{
		"currentUser": func() *user.User { return uPtr },
		"flashes":     func() []interface{} { return flashes },
		"currentUrl":  func() string { return currentUrl },
	}

	tmpl, err := template.New(templateName).Funcs(funcs).ParseFS(t.templates, files...)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func (t Templater) Render(w http.ResponseWriter, r *http.Request, data any, templateName string, files ...string) {
	err := t.renderErr(w, r, data, templateName, files...)
	if err != nil {
		t.logger.Error("error occurred during template rendering", "error", err)
		t.renderServerError(w, r)
	}
}

func (t Templater) RenderBase(w http.ResponseWriter, r *http.Request, data any, files ...string) {
	files = append(files, "base.gohtml")
	t.Render(w, r, data, "base", files...)
}

func (t Templater) renderServerError(w http.ResponseWriter, r *http.Request) {
	_ = t.renderErr(w, r, nil, "base", "base.gohtml", "errors/500.gohtml")
}

func getUserFromRequestContext(ctx context.Context) (user.User, error) {
	val := ctx.Value("user")
	if val == nil {
		return user.User{}, errors.New("no user set in request context")
	}

	u, ok := val.(user.User)
	if !ok {
		return user.User{}, errors.New("user set in request context is not valid")
	}

	return u, nil
}
