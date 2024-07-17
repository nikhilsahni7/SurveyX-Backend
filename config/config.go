package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
    GoogleOauthConfig *oauth2.Config
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Printf("Error loading .env file: %v", err)
    }

    GoogleOauthConfig = &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        RedirectURL:  "http://localhost:8080/auth/google/callback",
        Scopes: []string{
            "https://www.googleapis.com/auth/userinfo.email",
            "https://www.googleapis.com/auth/userinfo.profile",
        },
        Endpoint: google.Endpoint,
    }

    log.Printf("Google OAuth Config initialized with Client ID: %s", GoogleOauthConfig.ClientID)
}

func GenerateStateOauthCookie(w http.ResponseWriter) string {
    b := make([]byte, 16)
    rand.Read(b)
    state := base64.URLEncoding.EncodeToString(b)
    cookie := &http.Cookie{
        Name:     "oauthstate",
        Value:    state,
        Expires:  time.Now().Add(30 * time.Minute),
        HttpOnly: true,
        Path:     "/",
        SameSite: http.SameSiteLaxMode,
    }
    http.SetCookie(w, cookie)
    return state
}

func VerifyStateOauthCookie(r *http.Request) error {
    state := r.FormValue("state")
    cookie, err := r.Cookie("oauthstate")
    if err != nil {
        return err
    }
    if cookie.Value != state {
        return fmt.Errorf("invalid oauth state")
    }
    return nil
}
