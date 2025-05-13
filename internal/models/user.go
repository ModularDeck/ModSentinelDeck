package models

import (
	"time"
)

// Tenant represents an organization or company
type Tenant struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID        int       `json:"id"`
	TenantID  int       `json:"tenant_id"` // Reference to the tenant
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // do not expose password in API responses
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Team struct {
	ID          int       `json:"id"`
	TenantID    int       `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type UserTeam struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	TeamID    int       `json:"team_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Req_User_Login struct {
	UserName     string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	TenantName   string `json:"tenant_name,omitempty"`
	TenantDesc   string `json:"tenant_desc,omitempty"`
	TeamName     string `json:"team_name,omitempty"`
	TeamDesc     string `json:"team_desc,omitempty"`
	UserTeamRole string `json:"team_role,omitempty"`
	UserRole     string `json:"user_role,omitempty"`
}

// // HashPassword hashes the user's password
// func (u *User) HashPassword(password string) error {
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return err
// 	}
// 	u.Password = string(hashedPassword)
// 	return nil
// }

// // CheckPassword compares the hashed password with the plain-text password
// func (u *User) CheckPassword(password string) error {
// 	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
// }
