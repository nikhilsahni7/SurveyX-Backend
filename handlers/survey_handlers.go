package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
	"gorm.io/gorm"
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
	survey.Version = 1

	// Start a transaction
	tx := db.DB.Begin()

	if err := tx.Create(&survey).Error; err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range survey.Questions {
		survey.Questions[i].SurveyID = survey.ID
		survey.Questions[i].ID = 0 // Ensure we're creating new questions
		if err := tx.Create(&survey.Questions[i]).Error; err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for j := range survey.Questions[i].Options {
			survey.Questions[i].Options[j].QuestionID = survey.Questions[i].ID
			survey.Questions[i].Options[j].ID = 0 // Ensure we're creating new options
			if err := tx.Create(&survey.Questions[i].Options[j]).Error; err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		for j := range survey.Questions[i].Conditions {
			survey.Questions[i].Conditions[j].QuestionID = survey.Questions[i].ID
			survey.Questions[i].Conditions[j].ID = 0 // Ensure we're creating new conditions
			if err := tx.Create(&survey.Questions[i].Conditions[j]).Error; err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	link := models.SurveyLink{
		SurveyID: survey.ID,
		Link:     generateSurveyLink(survey.ID),
		IsActive: true,
	}

	if err := tx.Create(&link).Error; err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(surveys)
}

func GetSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var survey models.Survey
	if err := db.DB.Preload("Questions.Options").Preload("Questions.Conditions").First(&survey, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Survey not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(survey)
}

func UpdateSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var updatedSurvey models.Survey
	err = json.NewDecoder(r.Body).Decode(&updatedSurvey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingSurvey models.Survey
	if err := db.DB.Preload("Questions.Options").Preload("Questions.Conditions").First(&existingSurvey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	existingSurvey.Title = updatedSurvey.Title
	existingSurvey.Description = updatedSurvey.Description
	existingSurvey.Version++

	// Start a transaction
	tx := db.DB.Begin()

	// Update questions
	for i, updatedQuestion := range updatedSurvey.Questions {
		var existingQuestion models.Question
		if err := tx.Where("survey_id = ? AND id = ?", existingSurvey.ID, updatedQuestion.ID).First(&existingQuestion).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// This is a new question, create it
				updatedQuestion.SurveyID = existingSurvey.ID
				if err := tx.Create(&updatedQuestion).Error; err != nil {
					tx.Rollback()
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Update existing question
			existingQuestion.Text = updatedQuestion.Text
			existingQuestion.Type = updatedQuestion.Type
			existingQuestion.IsRequired = updatedQuestion.IsRequired
			existingQuestion.Order = updatedQuestion.Order
			existingQuestion.MinValue = updatedQuestion.MinValue
			existingQuestion.MaxValue = updatedQuestion.MaxValue
			existingQuestion.AllowMultiple = updatedQuestion.AllowMultiple
			existingQuestion.MaxFileSize = updatedQuestion.MaxFileSize

			if err := tx.Save(&existingQuestion).Error; err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Update options
			if err := tx.Where("question_id = ?", existingQuestion.ID).Delete(&models.Option{}).Error; err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, option := range updatedQuestion.Options {
				option.ID = 0 // Ensure we're creating new options
				option.QuestionID = existingQuestion.ID
				if err := tx.Create(&option).Error; err != nil {
					tx.Rollback()
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// Update conditions
			if err := tx.Where("question_id = ?", existingQuestion.ID).Delete(&models.Condition{}).Error; err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, condition := range updatedQuestion.Conditions {
				condition.ID = 0 // Ensure we're creating new conditions
				condition.QuestionID = existingQuestion.ID
				if err := tx.Create(&condition).Error; err != nil {
					tx.Rollback()
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		updatedSurvey.Questions[i] = updatedQuestion
	}

	// Delete questions that are no longer present
	var updatedQuestionIDs []uint
	for _, q := range updatedSurvey.Questions {
		updatedQuestionIDs = append(updatedQuestionIDs, q.ID)
	}
	if err := tx.Where("survey_id = ? AND id NOT IN ?", existingSurvey.ID, updatedQuestionIDs).Delete(&models.Question{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save the updated survey
	if err := tx.Save(&existingSurvey).Error; err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingSurvey)
}
func DeleteSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	if err := db.DB.Delete(&models.Survey{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DuplicateSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var originalSurvey models.Survey
	if err := db.DB.Preload("Questions.Options").Preload("Questions.Conditions").First(&originalSurvey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	newSurvey := originalSurvey
	newSurvey.ID = 0
	newSurvey.Title = "Copy of " + newSurvey.Title
	newSurvey.Version = 1
	newSurvey.IsPublished = false

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

		for j := range newSurvey.Questions[i].Conditions {
			newSurvey.Questions[i].Conditions[j].ID = 0
			newSurvey.Questions[i].Conditions[j].QuestionID = newSurvey.Questions[i].ID
			if err := db.DB.Create(&newSurvey.Questions[i].Conditions[j]).Error; err != nil {
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

	newSurvey.Link = link.Link

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newSurvey)
}

func PublishSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var survey models.Survey
	if err := db.DB.First(&survey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	survey.IsPublished = true
	if err := db.DB.Save(&survey).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(survey)
}

func UnpublishSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var survey models.Survey
	if err := db.DB.First(&survey, id).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	survey.IsPublished = false
	if err := db.DB.Save(&survey).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(survey)
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

	for i := range response.Answers {
		response.Answers[i].ID = 0
		response.Answers[i].ResponseID = response.ID
		if err := db.DB.Create(&response.Answers[i]).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	go TriggerWebhook(uint(surveyID), response.ID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func ListResponses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var responses []models.Response
	if err := db.DB.Where("survey_id = ?", surveyID).Find(&responses).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
func GetResponse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	responseID, err := strconv.ParseUint(vars["responseID"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid response ID", http.StatusBadRequest)
		return
	}

	var response models.Response
	if err := db.DB.Preload("Answers.Question.Options").First(&response, responseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Response not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

func generateSurveyLink(surveyID uint) string {
	return "survey-" + strconv.FormatUint(uint64(surveyID), 10)
}
