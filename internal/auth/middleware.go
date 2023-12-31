package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/OmBudhiraja/go-htmx-chat/internal/db"
)

type MiddlewareContextKey string

const UserContextKey MiddlewareContextKey = "user"

func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the token from the cookie
		token, err := r.Cookie(SessionCookieName)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		user, session, exists, err := db.GetUserAndSession(token.Value)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}

		if !exists {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		if time.Now().After(session.Expires) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Token Expired"))
			db.DeleteSession(token.Value)
			DeleteCookie(SessionCookieName)
			return
		}

		//// Calculate last updated date to throttle write updates to database
		// Formula: ({expiry date} - sessionMaxAge) + sessionUpdateAge
		//     e.g. ({expiry date} - 30 days) + 1 day
		sessionIsDueToBeUpdatedDate := session.Expires.Add(-db.SessionExpiry).Add(db.SessionUpdateAge)

		if time.Now().After(sessionIsDueToBeUpdatedDate) {
			fmt.Println("Updating session expiry")
			db.UpdateSessionExpiry(token.Value)
			http.SetCookie(w, CreateCookie(SessionCookieName, session.Token, time.Now().Add(db.SessionExpiry)))
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AttachUserToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the token from the cookie
		token, err := r.Cookie(SessionCookieName)

		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, _, exists, err := db.GetUserAndSession(token.Value)

		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
