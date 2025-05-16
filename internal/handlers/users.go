package handlers

import (
	"log"
	"sentinel/internal/auth"
	"sentinel/internal/db"
	"sentinel/internal/models"

	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUser handles user registration, along with tenant and team creation if needed
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.Req_User_Login
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" || req.Password == "" || req.TenantName == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Set default tenant name if not provided
	if req.TenantName == "" {
		req.TenantName = req.UserName + "'s Organization"
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		log.Println("Failed to start transaction:", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Step 1: Create or get tenant
	var tenantID int
	err = tx.QueryRow(`
		INSERT INTO tenants (name, description, created_at, updated_at) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description, updated_at = EXCLUDED.updated_at
		RETURNING id
	`, req.TenantName, req.TenantDesc, time.Now(), time.Now()).Scan(&tenantID)
	if err != nil {
		http.Error(w, "Error creating or updating tenant", http.StatusInternalServerError)
		return
	}

	// Step 2: Create or get user
	if req.UserRole == "" {
		req.UserRole = "member" // default role if not provided
	}
	var userID int
	err = tx.QueryRow(`
		INSERT INTO users (tenant_id, name, email, password, role, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (email) DO UPDATE SET password = EXCLUDED.password, updated_at = EXCLUDED.updated_at
		RETURNING id
	`, tenantID, req.UserName, req.Email, string(hashedPassword), req.UserRole, time.Now(), time.Now()).Scan(&userID)
	if err != nil {
		http.Error(w, "Error creating or updating user", http.StatusInternalServerError)
		return
	}

	// Step 3: Create or get team
	if req.TeamName == "" {
		req.TeamName = req.TenantName + " Default Team"
	}
	if req.TeamDesc == "" {
		req.TeamDesc = "Default team for " + req.TenantName
	}

	var teamID int
	err = tx.QueryRow(`
		INSERT INTO teams (tenant_id, name, description, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, name) 
		DO UPDATE SET description = EXCLUDED.description, updated_at = EXCLUDED.updated_at
		RETURNING id
	`, tenantID, req.TeamName, req.TeamDesc, time.Now(), time.Now()).Scan(&teamID)
	if err != nil {
		http.Error(w, "Error creating or updating team", http.StatusInternalServerError)
		return
	}

	// Step 4: Add user to team with default or provided role
	if req.UserTeamRole == "" {
		req.UserTeamRole = "member" // default role if not provided
	}

	_, err = tx.Exec(`
		INSERT INTO user_teams (user_id, team_id, role, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, team_id) DO UPDATE SET role = EXCLUDED.role, updated_at = EXCLUDED.updated_at
	`, userID, teamID, req.UserTeamRole, time.Now(), time.Now())
	if err != nil {
		http.Error(w, "Error adding user to team", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "User created successfully with tenant and team",
		"user_id":   strconv.Itoa(userID),
		"tenant_id": strconv.Itoa(tenantID),
		"team_id":   strconv.Itoa(teamID),
	})
}

// GetUserDetails fetches the user along with tenant and team information
func GetUserDetails(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id") // expecting ?id= in the URL query
	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch user, tenant, and team details
	var user models.User
	var tenant models.Tenant
	var team models.Team

	ctxTenantID, _ := auth.GetTenantID(r.Context())

	err = db.DB.QueryRow(`
		SELECT u.id, u.name, u.email, t.id, t.name, tm.id, tm.name
		FROM users u
		JOIN tenants t ON u.tenant_id = t.id
		LEFT JOIN user_teams ut ON ut.user_id = u.id
		LEFT JOIN teams tm ON tm.id = ut.team_id
		WHERE u.id = $1 and u.tenant_id = $2 `, userID, ctxTenantID).Scan(&user.ID, &user.Name, &user.Email, &tenant.ID, &tenant.Name, &team.ID, &team.Name)

	if err != nil {
		http.Error(w, "Error fetching user details", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":     user.ID,
		"user_name":   user.Name,
		"user_email":  user.Email,
		"tenant_id":   tenant.ID,
		"tenant_name": tenant.Name,
		"team_id":     team.ID,
		"team_name":   team.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUserDetails updates user, tenant, and team information
func UpdateUserDetails(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID     int    `json:"user_id"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		Password   string `json:"password,omitempty"`
		TenantName string `json:"tenant_name"`
		TeamName   string `json:"team_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Retrieve the user's email from the context
	ctx := r.Context()
	email, err := auth.GetEmail(ctx) // Assuming GetEmail returns email from context
	if err != nil || email == "" {
		http.Error(w, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Check if the user trying to update is the owner of the UserID
	var currentUser models.User
	err = db.DB.QueryRow("SELECT id, email, role FROM users WHERE id=$1", req.UserID).Scan(&currentUser.ID, &currentUser.Email, &currentUser.Role)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Allow update only if the user is an admin or is updating their own details
	if currentUser.Role != "admin" {
		http.Error(w, "Unauthorized to update this user", http.StatusUnauthorized)
		return
	}

	// Hash the new password if provided
	if req.Password != "" && req.Email != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error creating password", http.StatusInternalServerError)
			return
		}
		// Update user with password
		_, err = db.DB.Exec(`UPDATE users SET name=$1, email=$2, password=$3 WHERE id=$4`,
			req.Name, req.Email, hashedPassword, req.UserID)
		if err != nil {
			http.Error(w, "Error inserting password", http.StatusInternalServerError)
			return
		}
	} else {

		if req.Name != "" {
			_, err = db.DB.Exec(`UPDATE users SET name=$1 WHERE id=$2`,
				req.Name, req.UserID)
		}
	}
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	// Update tenant information
	if req.TenantName != "" {
		_, err = db.DB.Exec(`UPDATE tenants SET name=$1 WHERE id=(SELECT tenant_id FROM users WHERE id=$2)`, req.TenantName, req.UserID)
		if err != nil {
			http.Error(w, "Error updating tenant", http.StatusInternalServerError)
			return
		}
	}

	// Update or create a new team for the user
	if req.TeamName != "" {
		var teamID int
		err = db.DB.QueryRow(`SELECT id FROM teams WHERE name=$1 AND tenant_id=(SELECT tenant_id FROM users WHERE id=$2)`, req.TeamName, req.UserID).Scan(&teamID)
		if err == nil {
			// Team exists, update user_team relation
			_, err = db.DB.Exec(`UPDATE user_teams SET team_id=$1 WHERE user_id=$2`, teamID, req.UserID)
			if err != nil {
				http.Error(w, "Error updating user team", http.StatusInternalServerError)
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User details updated successfully"})
}
