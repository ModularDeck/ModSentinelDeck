
package user

import (
    "fmt"
    "sentinel/internal/utils"
)

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func RegisterUser(req RegisterRequest) error {
    hashedPassword, err := utils.HashPassword(req.Password)
    if err != nil {
        return err
    }

    // TODO: Store in database (mock for now)
    fmt.Printf("Registering user %s with hashed password: %s\n", req.Email, hashedPassword)
    return nil
}
