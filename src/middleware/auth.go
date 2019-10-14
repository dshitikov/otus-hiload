package middleware

import (
	"context"
	"github.com/alexedwards/scs/v2"
	"net/http"
	"otus-hiload/src/constants"
)

func AuthHandler(h http.Handler, sessionManager *scs.SessionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := sessionManager.GetBool(r.Context(), constants.CtxAuthenticated)

		if !auth {
			http.Redirect(w, r, constants.LoginPath, http.StatusFound)
			return
		}

		userID := sessionManager.Get(r.Context(), constants.CtxUserId).(int64)
		newRequest := r.WithContext(context.WithValue(r.Context(), constants.CtxUserId, userID))
		*r = *newRequest

		h.ServeHTTP(w, r)
	})
}
