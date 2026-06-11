package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/metrics"
	"github.com/ronexlemon/bnbcore/internal/senders"
)

type GoogleAuthRequest struct {
    Credential string `json:"credential"`
}

type RequestPasswordResetRequest struct {
    Email string `json:"email"`
}

type ResetPasswordRequest struct {
    Token       string `json:"token"`
    NewPassword string `json:"new_password"`
}

type UserHandler struct {
    Server         *http.ServeMux
    Service        *user.UserService
    JWTAuthManager *auth.JwtManager
    BaseUrl        string
	Stream         *eventstream.KafkaClient
    EmailSender  *senders.Sender
}
type RegisterRequest struct {
    Email     string `json:"email"`
    Password  string `json:"password"`
}
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func NewUserHandler(server *http.ServeMux, service *user.UserService, manager *auth.JwtManager, base string,stream  *eventstream.KafkaClient,email *senders.Sender) *UserHandler {
    h := &UserHandler{
        Server:         server,
        Service:        service,
        JWTAuthManager: manager,
        BaseUrl:        base,
		Stream: stream,
        EmailSender: email,
    }
    h.registerRoutes()
    return h
}

func (h *UserHandler) registerRoutes() {
    api := "/api/v1"
    h.Server.HandleFunc("POST "+api+"/users/register",metrics.MetricsMiddleware(h.CreateUser))
    h.Server.HandleFunc("GET  "+api+"/auth/verify",  metrics.MetricsMiddleware(h.VerifyMagicLink))      
h.Server.HandleFunc("POST "+api+"/auth/resend",     metrics.MetricsMiddleware(h.ResendVerification)) 
    h.Server.HandleFunc("POST "+api+"/users/login", metrics.MetricsMiddleware(h.LoginHandler))
    h.Server.HandleFunc("POST "+api+"/auth/google", metrics.MetricsMiddleware(h.GoogleAuth)) 
    h.Server.HandleFunc("POST "+api+"/auth/refresh", metrics.MetricsMiddleware(h.RefreshHandler))
    h.Server.HandleFunc("GET "+api+"/users/{id}", metrics.MetricsMiddleware(h.GetUser))
    h.Server.HandleFunc("POST "+api+"/auth/password-reset/request", metrics.MetricsMiddleware(h.RequestPasswordResetHandler))
    h.Server.HandleFunc("POST "+api+"/auth/password-reset/confirm", metrics.MetricsMiddleware(h.ResetPasswordHandler))
    h.Server.Handle("GET "+api+"/users/me", h.JWTAuthManager.Authenticate(http.HandlerFunc( metrics.MetricsMiddleware(h.GetMe))))
    h.Server.Handle("PUT "+api+"/users/{id}", h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.UpdateUser))))
    h.Server.Handle("DELETE "+api+"/users/{id}", h.JWTAuthManager.Authenticate(h.JWTAuthManager.RequireRole("owner")(http.HandlerFunc(metrics.MetricsMiddleware(h.DeleteUser)))))
}


func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body format", http.StatusBadRequest)
        return
    }

    usr, err := h.Service.Register(r.Context(), req.Email, req.Password)
    if err != nil {
        metrics.UserRegistrationsTotal.WithLabelValues("failure").Inc()
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    metrics.UserRegistrationsTotal.WithLabelValues("success").Inc()

    token, err := h.Service.CreateMagicLinkToken(r.Context(), usr.ID)
    if err != nil {
        http.Error(w, "could not create verification link", http.StatusInternalServerError)
        return
    }

    link := fmt.Sprintf("%s/register/verify-email?token=%s", h.BaseUrl, token)

    _ = h.Stream.Publish(r.Context(), eventstream.TopicUserSignedUp, usr.ID.String(),
        eventstream.UserSignedUpEvent{
            UserID:  usr.ID.String(),
            Email:   usr.Email,
            Name:     usr.Email,
            SignupLink: link,
        },
    )

	metrics.VerificationEmailsTotal.WithLabelValues("success").Inc()

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusAccepted)
writeJSON(w,map[string]string{"message":"check your email to complete sign-up"})

}

func (h *UserHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    if token == "" {
        http.Error(w, "missing token", http.StatusBadRequest)
        return
    }

    usr, err := h.Service.ValidateMagicLinkToken(r.Context(), token)
    if err != nil {
        fmt.Println("VERIFICATION ERROR",err)
        metrics.MagicLinksVerifiedTotal.WithLabelValues("invalid").Inc()
        http.Error(w, "invalid or expired link", http.StatusUnauthorized)
        return
    }
    metrics.MagicLinksVerifiedTotal.WithLabelValues("success").Inc()

    h.issueTokens(w, r, usr)
}

func (h *UserHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email string `json:"email"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body format", http.StatusBadRequest)
        return
    }

    usr, err := h.Service.GetUserByEmail(r.Context(), req.Email)
    if err != nil {
        w.WriteHeader(http.StatusAccepted)
        json.NewEncoder(w).Encode(map[string]string{
            "message": "if that email is registered, a new link is on its way",
        })
        return
    }

    if usr.IsActive {
        w.WriteHeader(http.StatusAccepted)
        json.NewEncoder(w).Encode(map[string]string{
            "message": "if that email is registered, a new link is on its way",
        })
        return
    }

    token, err := h.Service.CreateMagicLinkToken(r.Context(), usr.ID)
    if err != nil {
        http.Error(w, "could not create verification link", http.StatusInternalServerError)
        return
    }

    link := fmt.Sprintf("%s/register/verify-email?token=%s", h.BaseUrl, token)
    _ = h.Stream.Publish(r.Context(), eventstream.TopicUserSignedUp, usr.ID.String(),
        eventstream.UserSignedUpEvent{
            UserID:  usr.ID.String(),
            Email:   usr.Email,
            Name:     usr.Email,
            SignupLink: link,
        },
    )
metrics.MagicLinksIssuedTotal.WithLabelValues("resend").Inc()
	metrics.VerificationEmailsTotal.WithLabelValues("success").Inc()
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "if that email is registered, a new link is on its way",
    })
}
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        metrics.UserLoginsTotal.WithLabelValues("password", "failure").Inc()
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    userResult, err := h.Service.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        metrics.UserLoginsTotal.WithLabelValues("password", "failure").Inc()
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }
    if !userResult.IsActive {
        metrics.UserLoginsTotal.WithLabelValues("password", "unverified").Inc()
         token, err := h.Service.CreateMagicLinkToken(r.Context(), userResult.ID)
    if err != nil {
        http.Error(w, "could not create verification link", http.StatusInternalServerError)
        return
    }

    link := fmt.Sprintf("%s/register/verify-email?token=%s", h.BaseUrl, token)
    _ = h.Stream.Publish(r.Context(), eventstream.TopicUserSignedUp, userResult.ID.String(),
        eventstream.UserSignedUpEvent{
            UserID:  userResult.ID.String(),
            Email:   userResult.Email,
            Name:     userResult.Email,
            SignupLink: link,
        },
    )
    metrics.MagicLinksIssuedTotal.WithLabelValues("login_unverified").Inc()
		metrics.VerificationEmailsTotal.WithLabelValues("success").Inc()
        http.Error(w, "email not verified — check your inbox", http.StatusForbidden)
        return
    }
    metrics.UserLoginsTotal.WithLabelValues("password", "success").Inc()

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
    })
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

    _, newAccess, err := h.JWTAuthManager.RefreshAccessToken(body.RefreshToken, usr.Email, usr.Role)
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
        "data":usr,
    })
}

func (h *UserHandler) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()
    
    var req RequestPasswordResetRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body format", http.StatusBadRequest)
        return
    }

    if req.Email == "" {
        http.Error(w, "email is required", http.StatusBadRequest)
        return
    }

    err := h.Service.RequestPasswordReset(r.Context(), req.Email, h.BaseUrl)
    if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    writeJSON(w, map[string]string{"message": "if the email exists, a reset link has been sent"})
}

func (h *UserHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    var req ResetPasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body format", http.StatusBadRequest)
        return
    }

    if req.Token == "" || req.NewPassword == "" {
        http.Error(w, "token and new_password are required", http.StatusBadRequest)
        return
    }

    err := h.Service.ResetPassword(r.Context(), req.Token, req.NewPassword)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    writeJSON(w, map[string]string{"message": "password has been successfully updated"})
}