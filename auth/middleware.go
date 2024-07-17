package auth

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
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

	// Set session options
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true if using HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	log.Println("Session store initialized successfully")
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		// Even if there's an error, proceed to reset the cookie
	}

	// Reset session values
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1 // Expire the cookie immediately

	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session: %v", err)
		// If saving fails, manually expire the cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session-name",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})
	}
}
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "session-name")
		if err != nil {
			log.Printf("Error getting session: %v", err)
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"].(uint)
		if !ok {
			http.Error(w, "Invalid user session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
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
		GoogleID:     nil, // Explicitly set to nil for non-Google users
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
