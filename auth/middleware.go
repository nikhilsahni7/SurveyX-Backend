package auth

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/antonlindstrom/pgstore"
)

var (
	Store *pgstore.PGStore
)

func InitStore() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	Store, err = pgstore.NewPGStore(dsn, []byte(os.Getenv("SESSION_KEY")))
	if err != nil {
		log.Fatalf("Failed to initialize session store: %v", err)
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "session-name")
		if err != nil {
			http.Error(w, "Invalid session", http.StatusInternalServerError)
			return
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID := session.Values["user_id"].(uint)
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	session.Options.MaxAge = -1
	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	session.Save(r, w)
}
