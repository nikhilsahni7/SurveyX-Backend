package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
)

func CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook models.Webhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(uint)
	webhook.UserID = userID

	if err := db.DB.Create(&webhook).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(webhook)
}

func ListWebhooks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	var webhooks []models.Webhook

	if err := db.DB.Where("user_id = ?", userID).Find(&webhooks).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(webhooks)
}

func UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	var updatedWebhook models.Webhook
	err = json.NewDecoder(r.Body).Decode(&updatedWebhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var webhook models.Webhook
	if err := db.DB.First(&webhook, id).Error; err != nil {
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}

	webhook.URL = updatedWebhook.URL
	webhook.Events = updatedWebhook.Events
	webhook.Secret = updatedWebhook.Secret

	if err := db.DB.Save(&webhook).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(webhook)
}

func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	if err := db.DB.Delete(&models.Webhook{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func TriggerWebhook(surveyID uint, responseID uint) {
	var webhooks []models.Webhook
	db.DB.Where("survey_id = ?", surveyID).Find(&webhooks)

	for _, webhook := range webhooks {
		go func(hook models.Webhook) {
			payload := map[string]interface{}{
				"event":       "response_submitted",
				"survey_id":   surveyID,
				"response_id": responseID,
			}
			jsonPayload, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", hook.URL, bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Webhook-Secret", hook.Secret)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				// Log the error
				log.Printf("Error triggering webhook: %v", err)
			} else {
				defer resp.Body.Close()
				// Log the response status
				log.Printf("Webhook triggered. Status: %s", resp.Status)
			}
		}(webhook)
	}
}
