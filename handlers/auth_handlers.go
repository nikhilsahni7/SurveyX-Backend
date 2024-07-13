package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nikhilsahni7/SurveyX/auth"
	"github.com/nikhilsahni7/SurveyX/config"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoginHandler called. GoogleOauthConfig: %+v", config.GoogleOauthConfig)
	if config.GoogleOauthConfig.ClientID == "" || config.GoogleOauthConfig.ClientSecret == "" {
		log.Println("Error: Google OAuth ClientID or ClientSecret is empty")
		http.Error(w, "OAuth configuration error", http.StatusInternalServerError)
		return
	}

	// Generate a new state string for CSRF protection
	state := config.GenerateStateOauthCookie(w)
	url := config.GoogleOauthConfig.AuthCodeURL(state)
	log.Printf("Redirecting to Google OAuth URL: %s", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if err := config.VerifyStateOauthCookie(r, state); err != nil {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := config.GoogleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := auth.GetGoogleUserInfo(token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = auth.CreateOrUpdateUser(user)
	if err != nil {
		http.Error(w, "Failed to create/update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["user_id"] = user.ID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	var user models.User
	if err := db.GetDB().First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(
		map[string]string{
			"message": "Welcome to SurveyX " + user.Name,
		})
}

func TestAuthHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	var user models.User
	if err := db.GetDB().First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Authenticated as " + user.Email,
	})
}
