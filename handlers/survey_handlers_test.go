package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	dsn := "postgres://nikhilsahni:rajni.surender@localhost:5432/surveyxtest?sslmode=disable"
	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Auto Migrate the schema
	err = testDB.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.Survey{},
		&models.Question{},
		&models.Condition{},
		&models.Option{},
		&models.Response{},
		&models.Answer{},
		&models.SurveyLink{},
		&models.Webhook{},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate test database: %v", err))
	}

	return testDB
}

func TestSurveyHandlers(t *testing.T) {
	testDB := setupTestDB()
	db.DB = testDB
	defer func() {
		sqlDB, _ := testDB.DB()
		sqlDB.Close()
	}()

	router := mux.NewRouter()
	router.HandleFunc("/surveys", CreateSurvey).Methods("POST")
	router.HandleFunc("/surveys", ListSurveys).Methods("GET")
	router.HandleFunc("/surveys/{id}", GetSurvey).Methods("GET")
	router.HandleFunc("/surveys/{id}", UpdateSurvey).Methods("PUT")
	router.HandleFunc("/surveys/{id}", DeleteSurvey).Methods("DELETE")
	router.HandleFunc("/surveys/{id}/duplicate", DuplicateSurvey).Methods("POST")
	router.HandleFunc("/surveys/{id}/publish", PublishSurvey).Methods("POST")
	router.HandleFunc("/surveys/{id}/unpublish", UnpublishSurvey).Methods("POST")
	router.HandleFunc("/surveys/{id}/responses", SubmitResponse).Methods("POST")
	router.HandleFunc("/surveys/{id}/responses", ListResponses).Methods("GET")
	router.HandleFunc("/surveys/{id}/responses/{responseID}", GetResponse).Methods("GET")
	router.HandleFunc("/surveys/link/{linkID}", AccessSurveyByLink).Methods("GET")

	// Create a dummy user
	user := models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		GoogleID: nil,
	}
	db.DB.Create(&user)

	// Test CreateSurvey
	t.Run("CreateSurvey", func(t *testing.T) {
		survey := models.Survey{
			Title:       "Test Survey",
			Description: "This is a test survey",
			Questions: []models.Question{
				{
					Text: "What is your favorite color?",
					Type: "multiple_choice",
					Options: []models.Option{
						{Text: "Red", Value: "red"},
						{Text: "Blue", Value: "blue"},
						{Text: "Green", Value: "green"},
					},
					IsRequired: true,
					Order:      1,
				},
			},
		}

		body, _ := json.Marshal(survey)
		req, _ := http.NewRequest("POST", "/surveys", bytes.NewBuffer(body))
		req = req.WithContext(context.WithValue(req.Context(), "userID", user.ID))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var createdSurvey models.Survey
		json.Unmarshal(rr.Body.Bytes(), &createdSurvey)
		assert.NotEmpty(t, createdSurvey.ID)
		assert.Equal(t, survey.Title, createdSurvey.Title)
	})

	// Test ListSurveys
	t.Run("ListSurveys", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/surveys", nil)
		req = req.WithContext(context.WithValue(req.Context(), "userID", user.ID))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var surveys []models.Survey
		json.Unmarshal(rr.Body.Bytes(), &surveys)
		assert.NotEmpty(t, surveys)
	})

	// Test GetSurvey
	t.Run("GetSurvey", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Get",
			Description: "This is a test survey for get",
		}
		db.DB.Create(&survey)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/surveys/%d", survey.ID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var retrievedSurvey models.Survey
		json.Unmarshal(rr.Body.Bytes(), &retrievedSurvey)
		assert.Equal(t, survey.ID, retrievedSurvey.ID)
		assert.Equal(t, survey.Title, retrievedSurvey.Title)
	})

	// Test UpdateSurvey
	t.Run("UpdateSurvey", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Update",
			Description: "This is a test survey for update",
		}
		db.DB.Create(&survey)

		updatedSurvey := models.Survey{
			Title:       "Updated Test Survey",
			Description: "This is an updated test survey",
		}

		body, _ := json.Marshal(updatedSurvey)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/surveys/%d", survey.ID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var retrievedSurvey models.Survey
		json.Unmarshal(rr.Body.Bytes(), &retrievedSurvey)
		assert.Equal(t, survey.ID, retrievedSurvey.ID)
		assert.Equal(t, updatedSurvey.Title, retrievedSurvey.Title)
		assert.Equal(t, updatedSurvey.Description, retrievedSurvey.Description)
	})

	// Test DeleteSurvey
	t.Run("DeleteSurvey", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Delete",
			Description: "This is a test survey for delete",
		}
		db.DB.Create(&survey)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/surveys/%d", survey.ID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		var deletedSurvey models.Survey
		result := db.DB.First(&deletedSurvey, survey.ID)
		assert.Error(t, result.Error)
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
	})

	// Test DuplicateSurvey
	t.Run("DuplicateSurvey", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Duplicate",
			Description: "This is a test survey for duplicate",
		}
		db.DB.Create(&survey)

		req, _ := http.NewRequest("POST", fmt.Sprintf("/surveys/%d/duplicate", survey.ID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var duplicatedSurvey models.Survey
		json.Unmarshal(rr.Body.Bytes(), &duplicatedSurvey)
		assert.NotEqual(t, survey.ID, duplicatedSurvey.ID)
		assert.Equal(t, "Copy of "+survey.Title, duplicatedSurvey.Title)
	})

	// Test PublishSurvey
	t.Run("PublishSurvey", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Publish",
			Description: "This is a test survey for publish",
			IsPublished: false,
		}
		db.DB.Create(&survey)

		req, _ := http.NewRequest("POST", fmt.Sprintf("/surveys/%d/publish", survey.ID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var publishedSurvey models.Survey
		json.Unmarshal(rr.Body.Bytes(), &publishedSurvey)
		assert.True(t, publishedSurvey.IsPublished)
	})

	// Test UnpublishSurvey
	t.Run("UnpublishSurvey", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Unpublish",
			Description: "This is a test survey for unpublish",
			IsPublished: true,
		}
		db.DB.Create(&survey)

		req, _ := http.NewRequest("POST", fmt.Sprintf("/surveys/%d/unpublish", survey.ID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var unpublishedSurvey models.Survey
		json.Unmarshal(rr.Body.Bytes(), &unpublishedSurvey)
		assert.False(t, unpublishedSurvey.IsPublished)
	})

	// Test SubmitResponse
	t.Run("SubmitResponse", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Response",
			Description: "This is a test survey for response",
			Questions: []models.Question{
				{
					Text: "What is your favorite color?",
					Type: "multiple_choice",
					Options: []models.Option{
						{Text: "Red", Value: "red"},
						{Text: "Blue", Value: "blue"},
						{Text: "Green", Value: "green"},
					},
				},
			},
		}
		db.DB.Create(&survey)

		response := models.Response{
			Answers: []models.Answer{
				{
					QuestionID: survey.Questions[0].ID,
					Value:      "blue",
				},
			},
		}

		body, _ := json.Marshal(response)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/surveys/%d/responses", survey.ID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var submittedResponse models.Response
		json.Unmarshal(rr.Body.Bytes(), &submittedResponse)
		assert.NotEmpty(t, submittedResponse.ID)
		assert.Equal(t, survey.ID, submittedResponse.SurveyID)
	})

	// Test ListResponses
	t.Run("ListResponses", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for List Responses",
			Description: "This is a test survey for list responses",
		}
		db.DB.Create(&survey)

		for i := 0; i < 3; i++ {
			response := models.Response{
				SurveyID: survey.ID,
			}
			db.DB.Create(&response)
		}

		req, _ := http.NewRequest("GET", fmt.Sprintf("/surveys/%d/responses", survey.ID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responses []models.Response
		json.Unmarshal(rr.Body.Bytes(), &responses)
		assert.Len(t, responses, 3)
	})

	// Test GetResponse
	t.Run("GetResponse", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Get Response",
			Description: "This is a test survey for get response",
		}
		db.DB.Create(&survey)

		response := models.Response{
			SurveyID: survey.ID,
		}
		db.DB.Create(&response)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/surveys/%d/responses/%d", survey.ID, response.ID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var retrievedResponse models.Response
		json.Unmarshal(rr.Body.Bytes(), &retrievedResponse)
		assert.Equal(t, response.ID, retrievedResponse.ID)
		assert.Equal(t, survey.ID, retrievedResponse.SurveyID)
	})

	// Test AccessSurveyByLink
	t.Run("AccessSurveyByLink", func(t *testing.T) {
		survey := models.Survey{
			UserID:      user.ID,
			Title:       "Test Survey for Access By Link",
			Description: "This is a test survey for access by link",
		}
		db.DB.Create(&survey)

		link := models.SurveyLink{
			SurveyID: survey.ID,
			Link:     fmt.Sprintf("test-link-%d", survey.ID),
			IsActive: true,
		}
		db.DB.Create(&link)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/surveys/link/%s", link.Link), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var retrievedSurvey models.Survey
		json.Unmarshal(rr.Body.Bytes(), &retrievedSurvey)
		assert.Equal(t, survey.ID, retrievedSurvey.ID)
		assert.Equal(t, survey.Title, retrievedSurvey.Title)
	})
}

func setUserIDContext(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, "userID", userID)
}
