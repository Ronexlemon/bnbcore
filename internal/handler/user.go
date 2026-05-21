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

	var user  user.User

	err:=json.NewDecoder(r.Body).Decode(&user)
	if err !=nil{
		http.Error(w,"Invalid body",http.StatusBadRequest)
		return
	}

	user_result,err :=h.Service.Login(r.Context(),user.Email,user.PasswordHash)

	if err !=nil{
		http.Error(w,err.Error(),http.StatusNotFound)
		return
	}

	
		pair, err := h.JWTAuthManager.GenerateTokenPair(user_result.ID, user_result.Email, user_result.Role)
		if err != nil {
			http.Error(w, "could not generate tokens", http.StatusInternalServerError)
			return
		}

		h.Service.StoreRefreshToken(r.Context(),user_result.ID,pair.RefreshToken,pair.RefreshClaims.IssuedAt.Time,false,pair.RefreshClaims.ExpiresAt.Time)
		writeJSON(w, map[string]string{"access_token":  pair.AccessToken,
        "refresh_token": pair.RefreshToken,})
	
}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println("User id",id)

	w.WriteHeader(http.StatusNoContent) 
}


func (h *UserHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
    var body struct {
        RefreshToken string `json:"refresh_token"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
        http.Error(w, "refresh_token required", http.StatusBadRequest)
        return
    }

    tokenRecord, err := h.Service.GetRefreshToken(r.Context(), body.RefreshToken)
	
    user, err := h.Service.GetUserByID(r.Context(), tokenRecord.UserID)
    _,newAccess, err := h.JWTAuthManager.RefreshAccessToken(body.RefreshToken, user.Email, user.Role)
    if err != nil {
        http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
        return
    }
    writeJSON(w, map[string]string{"access_token": newAccess})
}