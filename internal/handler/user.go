package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
)


type UserHandler struct {
	Server *http.ServeMux
	Service *user.UserService
	JWTAuthManager *auth.JwtManager

}

func NewUserHandler(server *http.ServeMux,service *user.UserService,manager *auth.JwtManager) *UserHandler {
	h := &UserHandler{
		Server: server,
		Service: service,
		JWTAuthManager: manager,

	}
	h.registerRoutes()

	return h
}


func (h *UserHandler) registerRoutes() {
	api := "/api/v1"
	h.Server.HandleFunc("POST "+api+"/users", h.CreateUser)
	h.Server.HandleFunc("POST "+api+"/users/login", h.LoginHandler)
	h.Server.HandleFunc("GET "+api+"/users/{id}", h.GetUser)
	h.Server.HandleFunc("PUT "+api+"/users/{id}", h.UpdateUser)
	h.Server.HandleFunc("DELETE "+api+"/users/{id}", h.DeleteUser)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Fetching user with ID: " + id))
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Updated user with ID: " + id))
}

func(h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	
		pair, err := h.JWTAuthManager.GenerateTokenPair("user-123", "alice@example.com", "admin")
		if err != nil {
			http.Error(w, "could not generate tokens", http.StatusInternalServerError)
			return
		}
		writeJSON(w, pair)
	
}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println("User id",id)

	w.WriteHeader(http.StatusNoContent) 
}


func (h *UserHandler) refreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
			http.Error(w, "refresh_token required", http.StatusBadRequest)
			return
		}
 
		// TODO: look up fresh email/role from DB using the userID in
		// the refresh token claims before calling RefreshAccessToken.
		newAccess, err := h.JWTAuthManager.RefreshAccessToken(body.RefreshToken, "alice@example.com", "admin")
		if err != nil {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]string{"access_token": newAccess})
	}
}