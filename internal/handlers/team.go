package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sentinel/internal/auth"
	"sentinel/internal/db"
	"strconv"

	"github.com/gorilla/mux"
)

type TeamRequest struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TenantID    int    `json:"tenant_id"`
}

func CreateOrUpdateTeamHandler(w http.ResponseWriter, r *http.Request) {
	createOrUpdateTeam(w, r, db.DB) // Pass the global db instance here
}

func createOrUpdateTeam(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()
	role, err := auth.GetRole(ctx)
	if err != nil {
		http.Error(w, "Unauthorized: unable to determine role", http.StatusUnauthorized)
		return
	}
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		_, err := db.Exec(`INSERT INTO teams (name, description, tenant_id) VALUES ($1, $2, $3)`,
			req.Name, req.Description, req.TenantID)
		if err != nil {
			http.Error(w, "Error creating team", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "Team created successfully"}`))
	} else {
		_, err := db.Exec(`UPDATE teams SET name=$1, description=$2 WHERE id=$3`,
			req.Name, req.Description, req.ID)
		if err != nil {
			http.Error(w, "Error updating team", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message": "Team updated successfully"}`))
	}
}

func DeleteTeamHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role, _ := auth.GetRole(ctx)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	teamIDStr := vars["id"]
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Step 1: Update user_teams table to remove users from this team
	_, err = db.DB.Exec(`DELETE FROM user_teams WHERE team_id = $1`, teamID)
	if err != nil {
		http.Error(w, "Error dissociating users from team", http.StatusInternalServerError)
		return
	}

	// Step 2: Delete the team
	_, err = db.DB.Exec(`DELETE FROM teams WHERE id = $1`, teamID)
	if err != nil {
		http.Error(w, "Error deleting team", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Team deleted successfully"})
}
