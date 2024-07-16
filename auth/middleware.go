package auth

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/antonlindstrom/pgstore"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	Store *pgstore.PGStore
)

func InitStore() {
	var err error
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	Store, err = pgstore.NewPGStore(connStr, []byte(os.Getenv("SESSION_KEY")))
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateUser(email, name, password string) (*models.User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		Name:         name,
		PasswordHash: hashedPassword,
	}

	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
