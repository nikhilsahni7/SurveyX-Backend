package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
)

func GetSurveyAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var survey models.Survey
	if err := db.DB.Preload("Questions.Options").Preload("Responses.Answers").First(&survey, surveyID).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	analytics := calculateAnalytics(&survey)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func calculateAnalytics(survey *models.Survey) map[string]interface{} {
	analytics := make(map[string]interface{})

	totalResponses := len(survey.Responses)
	analytics["totalResponses"] = totalResponses

	questionAnalytics := make(map[string]interface{})
	for _, question := range survey.Questions {
		qa := make(map[string]interface{})

		switch question.Type {
		case "multipleChoice", "checkbox", "dropdown":
			optionCounts := make(map[string]int)
			for _, response := range survey.Responses {
				for _, answer := range response.Answers {
					if answer.QuestionID == question.ID {
						optionCounts[answer.Value]++
					}
				}
			}
			qa["optionCounts"] = optionCounts
		case "rating", "scale":
			var sum, count int
			for _, response := range survey.Responses {
				for _, answer := range response.Answers {
					if answer.QuestionID == question.ID {
						value, _ := strconv.Atoi(answer.Value)
						sum += value
						count++
					}
				}
			}
			if count > 0 {
				qa["average"] = float64(sum) / float64(count)
			}
		case "text", "textarea":
			answers := []string{}
			for _, response := range survey.Responses {
				for _, answer := range response.Answers {
					if answer.QuestionID == question.ID {
						answers = append(answers, answer.Value)
					}
				}
			}
			qa["answers"] = answers
		}

		questionAnalytics[strconv.Itoa(int(question.ID))] = qa
	}

	analytics["questionAnalytics"] = questionAnalytics

	return analytics
}

func ExportSurveyData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}

	var survey models.Survey
	if err := db.DB.Preload("Questions").Preload("Responses.Answers").First(&survey, surveyID).Error; err != nil {
		http.Error(w, "Survey not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=survey_data.csv")

	csvWriter := csv.NewWriter(w)

	// Write header
	header := []string{"ResponseID", "Timestamp"}
	for _, question := range survey.Questions {
		header = append(header, question.Text)
	}
	csvWriter.Write(header)

	// Write data
	for _, response := range survey.Responses {
		row := []string{strconv.Itoa(int(response.ID)), response.CreatedAt.String()}
		answerMap := make(map[uint]string)
		for _, answer := range response.Answers {
			answerMap[answer.QuestionID] = answer.Value
		}
		for _, question := range survey.Questions {
			row = append(row, answerMap[question.ID])
		}
		csvWriter.Write(row)
	}

	csvWriter.Flush()
}
