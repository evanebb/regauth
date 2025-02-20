package middleware

import (
	"net/http"
	"strings"
)

func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			method := strings.ToUpper(r.PostFormValue("_method"))

			if method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
				r.Method = method
			}
		}

		next.ServeHTTP(w, r)
	})
}
