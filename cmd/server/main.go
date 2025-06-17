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
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.LoginHandler(w, r, db.DB)
	}).Methods("POST")
	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST") // Add this for logout

	// Secure routes with JWT middleware
	secure := r.PathPrefix("/api").Subrouter()
	// Apply rate limiting to public routes
	r.Use(auth.RateLimitMiddleware)
	secure.Use(func(next http.Handler) http.Handler {
		return auth.AuthMiddleware(next, db.DB) // Wrap AuthMiddleware to match mux.MiddlewareFunc
	})
	secure.HandleFunc("/user/{id}", handlers.GetUserDetails).Methods("GET")
	secure.HandleFunc("/user/tenant/{tenant_id}", handlers.GetUsersByTenant).Methods("GET")

	secure.HandleFunc("/user", handlers.UpdateUserDetailsHandler).Methods("PUT")
	log.Println("Routers End")
	log.Println("Sentinel starting on :8080")
	handler := auth.EnableCORS(r) // ✅ wrap router with CORS middleware
	log.Fatal(http.ListenAndServe(":8080", handler))

}
