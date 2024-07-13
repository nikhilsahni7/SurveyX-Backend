package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nikhilsahni7/SurveyX/auth"
	"github.com/nikhilsahni7/SurveyX/db"
	"github.com/nikhilsahni7/SurveyX/handlers"
	"github.com/rs/cors"
)

func main() {
	db.InitDB()
	auth.InitStore()
	r := mux.NewRouter()

	// cors Middleware

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	// Auth routes
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	r.HandleFunc("/logout", handlers.LogoutHandler)
	r.HandleFunc("/api/test-auth", auth.AuthMiddleware(handlers.TestAuthHandler))

	// Protected routes
	r.HandleFunc("/", auth.AuthMiddleware(handlers.HomeHandler))
	r.HandleFunc("/dashboard", auth.AuthMiddleware(handlers.DashboardHandler))
	r.HandleFunc("/api/surveys", auth.AuthMiddleware(handlers.CreateSurvey)).Methods("POST")
	r.HandleFunc("/api/surveys", auth.AuthMiddleware(handlers.ListSurveys)).Methods("GET")
	r.HandleFunc("/api/surveys/{id}", auth.AuthMiddleware(handlers.GetSurvey)).Methods("GET")
	r.HandleFunc("/api/surveys/{id}/responses", auth.AuthMiddleware(handlers.GetSurveyResponses)).Methods("GET")
	r.HandleFunc("/api/surveys/{id}/submit", handlers.SubmitResponse).Methods("POST")
	r.HandleFunc("/s/{linkID}", handlers.AccessSurveyByLink).Methods("GET")

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
