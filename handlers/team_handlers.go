package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
)

func CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team models.Team
	err := json.NewDecoder(r.Body).Decode(&team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(uint)
	team.OwnerID = userID

	if err := db.DB.Create(&team).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func ListTeams(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	var teams []models.Team

	if err := db.DB.Where("owner_id = ?", userID).Or("id IN (?)", db.DB.Table("user_teams").Select("team_id").Where("user_id = ?", userID)).Find(&teams).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(teams)
}

func GetTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["teamId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var team models.Team
	if err := db.DB.Preload("Users").First(&team, teamID).Error; err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(team)
}

func UpdateTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["teamId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var updatedTeam models.Team
	err = json.NewDecoder(r.Body).Decode(&updatedTeam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var team models.Team
	if err := db.DB.First(&team, teamID).Error; err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	team.Name = updatedTeam.Name

	if err := db.DB.Save(&team).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(team)
}

func AddTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["teamId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Email string `json:"email"`
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var team models.Team
	if err := db.DB.First(&team, teamID).Error; err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := db.DB.Model(&team).Association("Users").Append(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User added to team successfully"})
}

func RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["teamId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(vars["userId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var team models.Team
	if err := db.DB.First(&team, teamID).Error; err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := db.DB.Model(&team).Association("Users").Delete(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User removed from team successfully"})
}
