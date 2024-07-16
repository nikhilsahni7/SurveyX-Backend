package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/auth"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/handlers"
	"github.com/nikhilsahni7/SurveyX/middlewares"
	"github.com/rs/cors"
)

func main() {
	db.InitDB()
	auth.InitStore()

	r := mux.NewRouter()

	// CORS Middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Middleware
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.RecoverMiddleware)

	// Rate limiting
	limiter := middlewares.NewIPRateLimiter(1, 5)
	r.Use(middlewares.LimitMiddleware(limiter))

	// Auth routes
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandlerEmail).Methods("POST")
	r.HandleFunc("/login/google", handlers.LoginHandler)
	r.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	r.HandleFunc("/logout", handlers.LogoutHandler)
	r.HandleFunc("/api/test-auth", auth.AuthMiddleware(handlers.TestAuthHandler))
	r.HandleFunc("/api/user", auth.AuthMiddleware(handlers.GetCurrentUser)).Methods("GET")

	// Survey routes
	r.HandleFunc("/api/surveys", auth.AuthMiddleware(handlers.CreateSurvey)).Methods("POST")
	r.HandleFunc("/api/surveys", auth.AuthMiddleware(handlers.ListSurveys)).Methods("GET")
	r.HandleFunc("/api/surveys/{id}", auth.AuthMiddleware(handlers.GetSurvey)).Methods("GET")
	r.HandleFunc("/api/surveys/{id}", auth.AuthMiddleware(handlers.UpdateSurvey)).Methods("PUT")
	r.HandleFunc("/api/surveys/{id}", auth.AuthMiddleware(handlers.DeleteSurvey)).Methods("DELETE")
	r.HandleFunc("/api/surveys/{id}/duplicate", auth.AuthMiddleware(handlers.DuplicateSurvey)).Methods("POST")
	r.HandleFunc("/api/surveys/{id}/publish", auth.AuthMiddleware(handlers.PublishSurvey)).Methods("POST")
	r.HandleFunc("/api/surveys/{id}/unpublish", auth.AuthMiddleware(handlers.UnpublishSurvey)).Methods("POST")

	// Response routes
	r.HandleFunc("/api/surveys/{id}/submit", handlers.SubmitResponse).Methods("POST")
	r.HandleFunc("/api/surveys/{id}/responses", auth.AuthMiddleware(handlers.ListResponses)).Methods("GET")
	r.HandleFunc("/api/surveys/{id}/responses/{responseId}", auth.AuthMiddleware(handlers.GetResponse)).Methods("GET")

	// Public survey access
	r.HandleFunc("/api/s/{linkID}", handlers.AccessSurveyByLink).Methods("GET")

	// Analytics routes
	r.HandleFunc("/api/surveys/{id}/analytics", auth.AuthMiddleware(handlers.GetSurveyAnalytics)).Methods("GET")
	r.HandleFunc("/api/surveys/{id}/export", auth.AuthMiddleware(handlers.ExportSurveyData)).Methods("GET")

	// Team routes
	r.HandleFunc("/api/teams", auth.AuthMiddleware(handlers.CreateTeam)).Methods("POST")
	r.HandleFunc("/api/teams", auth.AuthMiddleware(handlers.ListTeams)).Methods("GET")
	r.HandleFunc("/api/teams/{teamId}", auth.AuthMiddleware(handlers.GetTeam)).Methods("GET")
	r.HandleFunc("/api/teams/{teamId}", auth.AuthMiddleware(handlers.UpdateTeam)).Methods("PUT")
	r.HandleFunc("/api/teams/{teamId}/members", auth.AuthMiddleware(handlers.AddTeamMember)).Methods("POST")
	r.HandleFunc("/api/teams/{teamId}/members/{userId}", auth.AuthMiddleware(handlers.RemoveTeamMember)).Methods("DELETE")

	// Webhook routes
	r.HandleFunc("/api/webhooks", auth.AuthMiddleware(handlers.CreateWebhook)).Methods("POST")
	r.HandleFunc("/api/webhooks", auth.AuthMiddleware(handlers.ListWebhooks)).Methods("GET")
	r.HandleFunc("/api/webhooks/{id}", auth.AuthMiddleware(handlers.UpdateWebhook)).Methods("PUT")
	r.HandleFunc("/api/webhooks/{id}", auth.AuthMiddleware(handlers.DeleteWebhook)).Methods("DELETE")

	handler := c.Handler(r)

	srv := &http.Server{
		Handler:      handler,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Server starting on :8080")
	log.Fatal(srv.ListenAndServe())
}
