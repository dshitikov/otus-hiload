package middleware

import (
	"context"
	"github.com/gorilla/sessions"
	"net/http"
	"otus-hiload/src/constants"
)

func AuthHandler(h http.Handler, store *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, constants.CookieName)

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, constants.LoginPath, http.StatusFound)
			return
		}

		newRequest := r.WithContext(context.WithValue(r.Context(), constants.CtxUserId, session.Values["user_id"]))
		*r = *newRequest

		h.ServeHTTP(w, r)
	})
}
