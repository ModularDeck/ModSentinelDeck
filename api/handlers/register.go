
package handlers

import (
    "encoding/json"
    "net/http"
    "sentinel/pkg/user"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    var req user.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    if err := user.RegisterUser(req); err != nil {
        http.Error(w, "Could not register user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}
