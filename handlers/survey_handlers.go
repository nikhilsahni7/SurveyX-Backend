package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
)

func CreateSurvey(w http.ResponseWriter, r *http.Request) {
	var survey models.Survey
	err := json.NewDecoder(r.Body).Decode(&survey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(uint)
	survey.UserID = userID

	if err := db.DB.Create(&survey).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	link := models.SurveyLink{
		SurveyID: survey.ID,
		Link:     generateSurveyLink(survey.ID),
		IsActive: true,
	}

	if err := db.DB.Create(&link).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	survey.Link = link.Link

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(survey)
}

func ListSurveys(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	var surveys []models.Survey
	if err := db.DB.Where("user_id = ?", userID).Find(&surveys).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(surveys)
}

func GetSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var survey models.Survey
	if err := db.DB.Preload("Questions.Options").First(&survey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(survey)
}

func GetSurveyResponses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var responses []models.Response
	if err := db.DB.Where("survey_id = ?", id).Preload("Answers").Find(&responses).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(responses)
}

func SubmitResponse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var response models.Response
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response.SurveyID = uint(surveyID)
	response.IP = r.RemoteAddr
	response.UserAgent = r.UserAgent()

	if err := db.DB.Create(&response).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func AccessSurveyByLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["linkID"]
	var surveyLink models.SurveyLink
	if err := db.DB.Where("link = ? AND is_active = ?", linkID, true).First(&surveyLink).Error; err != nil {
		http.Error(w, "Survey not found or inactive", http.StatusNotFound)
		return
	}

	var survey models.Survey
	if err := db.DB.Preload("Questions.Options").First(&survey, surveyLink.SurveyID).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(survey)
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Write([]byte("Welcome to your dashboard, " + user.Name + "!"))
}

func generateSurveyLink(surveyID uint) string {
	return "survey-" + strconv.FormatUint(uint64(surveyID), 10)
}
