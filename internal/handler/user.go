package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
)

type GoogleAuthRequest struct {
    Credential string `json:"credential"`
    ShopName   string `json:"shop_name"` 
    Subdomain  string `json:"subdomain"` 
}

type UserHandler struct {
    Server         *http.ServeMux
    Service        *user.UserService
    JWTAuthManager *auth.JwtManager
    BaseUrl        string
	Stream         *eventstream.KafkaClient
}
type RegisterRequest struct {
    Email     string `json:"email"`
    Password  string `json:"password"`
}

func NewUserHandler(server *http.ServeMux, service *user.UserService, manager *auth.JwtManager, base string,stream  *eventstream.KafkaClient) *UserHandler {
    h := &UserHandler{
        Server:         server,
        Service:        service,
        JWTAuthManager: manager,
        BaseUrl:        base,
		Stream: stream,
    }
    h.registerRoutes()
    return h
}

func (h *UserHandler) registerRoutes() {
    api := "/api/v1"
    h.Server.HandleFunc("POST "+api+"/users/register", h.CreateUser)
    h.Server.HandleFunc("POST "+api+"/users/login", h.LoginHandler)
    h.Server.HandleFunc("POST "+api+"/auth/google", h.GoogleAuth) 
    h.Server.HandleFunc("POST "+api+"/auth/refresh", h.RefreshHandler)
    h.Server.HandleFunc("GET "+api+"/users/{id}", h.GetUser)
    h.Server.Handle("GET "+api+"/users/me", h.JWTAuthManager.Authenticate(http.HandlerFunc(h.GetMe)))
    h.Server.Handle("PUT "+api+"/users/{id}", h.JWTAuthManager.Authenticate(http.HandlerFunc(h.UpdateUser)))
    h.Server.Handle("DELETE "+api+"/users/{id}", h.JWTAuthManager.Authenticate(h.JWTAuthManager.RequireRole("owner")(http.HandlerFunc(h.DeleteUser))))
}


func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body format", http.StatusBadRequest)
        return
    }



    usr, err := h.Service.Register(r.Context(), req.Email, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	h.issueTokens(w, r, usr)
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
    var req user.User
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    userResult, err := h.Service.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    h.issueTokens(w, r, userResult)
}

func (h *UserHandler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
    var req GoogleAuthRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    if req.Credential == "" {
        http.Error(w, "credential is required", http.StatusBadRequest)
        return
    }

    userResult, err := h.Service.RegisterWithGoogle(r.Context(), user.GoogleLoginRequest{
        Credential: req.Credential,
    }, req.ShopName, req.Subdomain)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    h.issueTokens(w, r, userResult)
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
    if err != nil {
        http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
        return
    }

    usr, err := h.Service.GetUserByID(r.Context(), *tokenRecord.UserID)
    if err != nil {
        http.Error(w, "user not found", http.StatusUnauthorized)
        return
    }

    _, newAccess, err := h.JWTAuthManager.RefreshAccessToken(body.RefreshToken, usr.Email, usr.Role, usr.Subdomain)
    if err != nil {
        http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
        return
    }

    writeJSON(w, map[string]string{"access_token": newAccess})
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

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Println("User id", id)
    w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

	if claims.UserID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
userID := *claims.UserID

    userResult, err := h.Service.GetUserByID(r.Context(), userID)
    if err != nil {
        http.Error(w, "user not found", http.StatusNotFound)
        return
    }

    writeJSON(w, map[string]any{"user": userResult})
}

func (h *UserHandler) issueTokens(w http.ResponseWriter, r *http.Request, usr *user.User) {
    pair, err := h.JWTAuthManager.GenerateTokenPair(&usr.ID,usr.Email, usr.Role)
    if err != nil {
        http.Error(w, "could not generate tokens", http.StatusInternalServerError)
        return
    }

    if err := h.Service.StoreRefreshToken(
        r.Context(),
        usr.ID,
        pair.RefreshToken,
        pair.RefreshClaims.IssuedAt.Time,
        false,
        pair.RefreshClaims.ExpiresAt.Time,
    ); err != nil {
        http.Error(w, "could not store refresh token", http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]any{
        "access_token":  pair.AccessToken,
        "refresh_token": pair.RefreshToken,
    })
}

