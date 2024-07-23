package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
	"gorm.io/gorm"
)

func CreateSurvey(w http.ResponseWriter, r *http.Request) {
	var survey models.Survey
	if err := json.NewDecoder(r.Body).Decode(&survey); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(uint)
	survey.UserID = userID
	survey.Version = 1
	survey.ReleaseDate = withDefaultTime(survey.ReleaseDate, time.Now())
	survey.CloseDate = withDefaultTime(survey.CloseDate, time.Now().AddDate(0, 1, 0))

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Create the survey
		if err := tx.Create(&survey).Error; err != nil {
			return err
		}

		// Create questions and options
		for i := range survey.Questions {
			survey.Questions[i].ID = 0 // Ensure new record is created
			survey.Questions[i].SurveyID = survey.ID
			if err := tx.Create(&survey.Questions[i]).Error; err != nil {
				return err
			}

			for j := range survey.Questions[i].Options {
				survey.Questions[i].Options[j].ID = 0 // Ensure new record is created
				survey.Questions[i].Options[j].QuestionID = survey.Questions[i].ID
				if err := tx.Create(&survey.Questions[i].Options[j]).Error; err != nil {
					return err
				}
			}
		}

		// Create survey link
		link := models.SurveyLink{
			SurveyID: survey.ID,
			Link:     generateSurveyLink(survey.ID),
			IsActive: true,
		}
		if err := tx.Create(&link).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch the created survey with all its relations
	var createdSurvey models.Survey
	if err := db.DB.Preload("Questions.Options").First(&createdSurvey, survey.ID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSurvey)
}
func UpdateSurvey(w http.ResponseWriter, r *http.Request) {
	id := parseUintParam(r, "id")

	var updatedSurvey models.Survey
	if err := json.NewDecoder(r.Body).Decode(&updatedSurvey); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingSurvey models.Survey
	if err := db.DB.First(&existingSurvey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	existingSurvey.Title = updatedSurvey.Title
	existingSurvey.Description = updatedSurvey.Description
	existingSurvey.Version++

	if err := db.DB.Save(&existingSurvey).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := db.DB.Where("survey_id = ?", id).Delete(&models.Question{}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range updatedSurvey.Questions {
		updatedSurvey.Questions[i].ID = 0
		updatedSurvey.Questions[i].SurveyID = existingSurvey.ID
		if err := db.DB.Create(&updatedSurvey.Questions[i]).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for j := range updatedSurvey.Questions[i].Options {
			updatedSurvey.Questions[i].Options[j].ID = 0
			updatedSurvey.Questions[i].Options[j].QuestionID = updatedSurvey.Questions[i].ID
			if err := db.DB.Create(&updatedSurvey.Questions[i].Options[j]).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	json.NewEncoder(w).Encode(existingSurvey)
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
	id := parseUintParam(r, "id")

	var survey models.Survey
	if err := db.DB.Preload("Questions.Options").First(&survey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(survey)
}

func DeleteSurvey(w http.ResponseWriter, r *http.Request) {
	id := parseUintParam(r, "id")

	if err := db.DB.Delete(&models.Survey{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func PublishSurvey(w http.ResponseWriter, r *http.Request) {
	updateSurveyStatus(w, r, true)
}

func UnpublishSurvey(w http.ResponseWriter, r *http.Request) {
	updateSurveyStatus(w, r, false)
}

func updateSurveyStatus(w http.ResponseWriter, r *http.Request, isPublished bool) {
	id := parseUintParam(r, "id")

	if err := db.DB.Model(&models.Survey{}).Where("id = ?", id).Update("is_published", isPublished).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	GetSurvey(w, r)
}

func SubmitResponse(w http.ResponseWriter, r *http.Request) {
	surveyID := parseUintParam(r, "id")

	var responseData struct {
		Answers []struct {
			QuestionID uint   `json:"questionId"`
			Value      string `json:"value"`
		} `json:"answers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&responseData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := models.Response{
		SurveyID:  surveyID,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}

	if err := db.DB.Create(&response).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, answerData := range responseData.Answers {
		answer := models.Answer{
			ResponseID: response.ID,
			QuestionID: answerData.QuestionID,
			Value:      answerData.Value,
		}
		if err := db.DB.Create(&answer).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Response submitted successfully"})
}

func ListResponses(w http.ResponseWriter, r *http.Request) {
	surveyID := parseUintParam(r, "id")

	var responses []models.Response
	if err := db.DB.Where("survey_id = ?", surveyID).Preload("Answers").Find(&responses).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(responses)
}

func DuplicateSurvey(w http.ResponseWriter, r *http.Request) {
	id := parseUintParam(r, "id")

	var originalSurvey models.Survey
	if err := db.DB.Preload("Questions.Options").First(&originalSurvey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	newSurvey := originalSurvey
	newSurvey.ID = 0
	newSurvey.Title = "Copy of " + newSurvey.Title
	newSurvey.Version = 1
	newSurvey.IsPublished = false
	newSurvey.CreatedAt = time.Now()
	newSurvey.UpdatedAt = time.Now()

	if err := db.DB.Create(&newSurvey).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range newSurvey.Questions {
		newSurvey.Questions[i].ID = 0
		newSurvey.Questions[i].SurveyID = newSurvey.ID
		if err := db.DB.Create(&newSurvey.Questions[i]).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for j := range newSurvey.Questions[i].Options {
			newSurvey.Questions[i].Options[j].ID = 0
			newSurvey.Questions[i].Options[j].QuestionID = newSurvey.Questions[i].ID
			if err := db.DB.Create(&newSurvey.Questions[i].Options[j]).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	link := models.SurveyLink{
		SurveyID: newSurvey.ID,
		Link:     generateSurveyLink(newSurvey.ID),
		IsActive: true,
	}
	if err := db.DB.Create(&link).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newSurvey)
}

func AccessSurveyByLink(w http.ResponseWriter, r *http.Request) {
	linkID := mux.Vars(r)["linkID"]

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

	// Remove sensitive information
	survey.UserID = 0
	survey.Responses = nil

	json.NewEncoder(w).Encode(survey)
}

func GetResponse(w http.ResponseWriter, r *http.Request) {
	surveyID := parseUintParam(r, "id")
	responseID := parseUintParam(r, "responseId")

	var response models.Response
	if err := db.DB.Where("survey_id = ? AND id = ?", surveyID, responseID).Preload("Answers").First(&response).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Response not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Optionally, you can load the questions to provide more context
	var survey models.Survey
	if err := db.DB.Preload("Questions").First(&survey, surveyID).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	// Create a more informative response structure
	type AnswerWithQuestion struct {
		QuestionID   uint   `json:"questionId"`
		QuestionText string `json:"questionText"`
		Value        string `json:"value"`
	}

	type ResponseWithQuestions struct {
		ID        uint                 `json:"id"`
		SurveyID  uint                 `json:"surveyId"`
		CreatedAt time.Time            `json:"createdAt"`
		IP        string               `json:"ip"`
		UserAgent string               `json:"userAgent"`
		Answers   []AnswerWithQuestion `json:"answers"`
	}

	responseWithQuestions := ResponseWithQuestions{
		ID:        response.ID,
		SurveyID:  response.SurveyID,
		CreatedAt: response.CreatedAt,
		IP:        response.IP,
		UserAgent: response.UserAgent,
		Answers:   make([]AnswerWithQuestion, 0),
	}

	questionMap := make(map[uint]string)
	for _, question := range survey.Questions {
		questionMap[question.ID] = question.Text
	}

	for _, answer := range response.Answers {
		responseWithQuestions.Answers = append(responseWithQuestions.Answers, AnswerWithQuestion{
			QuestionID:   answer.QuestionID,
			QuestionText: questionMap[answer.QuestionID],
			Value:        answer.Value,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseWithQuestions)
}

// Helper functions

func withDefaultTime(t *time.Time, defaultTime time.Time) *time.Time {
	if t == nil {
		return &defaultTime
	}
	return t
}

func generateSurveyLink(surveyID uint) string {
	return fmt.Sprintf("survey-%d", surveyID)
}

func parseUintParam(r *http.Request, key string) uint {
	value, _ := strconv.ParseUint(mux.Vars(r)[key], 10, 64)
	return uint(value)
}
