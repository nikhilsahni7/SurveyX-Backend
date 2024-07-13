package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func GetGoogleUserInfo(token string) (*models.User, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer resp.Body.Close()

	var googleUser GoogleUserInfo
	if err = json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed decoding user info: %s", err.Error())
	}

	user := &models.User{
		GoogleID: googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Picture:  googleUser.Picture,
	}

	return user, nil
}

func CreateOrUpdateUser(user *models.User) error {
	var existingUser models.User
	result := db.DB.Where("google_id = ?", user.GoogleID).First(&existingUser)
	if result.Error != nil {
		// User doesn't exist, create new user
		if err := db.DB.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %s", err.Error())
		}
	} else {
		// User exists, update information
		existingUser.Name = user.Name
		existingUser.Email = user.Email
		if err := db.DB.Save(&existingUser).Error; err != nil {
			return fmt.Errorf("failed to update user: %s", err.Error())
		}
		*user = existingUser // Update the user object with database values
	}
	return nil
}
