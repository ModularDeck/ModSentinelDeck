package main

import (
	"log"
	"net/http"

	"sentinel/internal/auth"
	"sentinel/internal/db"
	"sentinel/internal/handlers"

	"github.com/gorilla/mux"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Initialize the database
	log.Println("Init DB")

	db.InitDB()
	defer db.DB.Close()
	log.Println("Defere DB")

	// Create a new router
	r := mux.NewRouter()
	log.Println("Routers Start")

	// Health check endpoint
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// Public routes
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST") // Add this for logout

	// Apply rate limiting to public routes
	r.Use(auth.RateLimitMiddleware)

	// Secure routes with JWT middleware
	secure := r.PathPrefix("/api").Subrouter()
	secure.Use(auth.AuthMiddleware)
	secure.HandleFunc("/user", handlers.GetUserDetails).Methods("GET")
	secure.HandleFunc("/user", handlers.UpdateUserDetails).Methods("PUT")
	log.Println("Routers End")
	log.Println("Sentinel starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
