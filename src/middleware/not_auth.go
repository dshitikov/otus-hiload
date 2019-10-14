package middleware

import (
	"github.com/alexedwards/scs/v2"
	"net/http"
	"otus-hiload/src/constants"
)

func NotAuthHandler(h http.Handler, sessionManager *scs.SessionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := sessionManager.GetBool(r.Context(), constants.CtxAuthenticated)

		if auth {
			http.Redirect(w, r, constants.RootPath, http.StatusFound)
			return
		}

		h.ServeHTTP(w, r)
	})
}
