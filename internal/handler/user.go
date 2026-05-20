package handler

import (
	"fmt"
	"net/http"

	"github.com/ronexlemon/bnbcore/internal/domain/user"
)


type UserHandler struct {
	Server *http.ServeMux
	Service *user.UserService

}

func NewUserHandler(server *http.ServeMux,service *user.UserService) *UserHandler {
	h := &UserHandler{
		Server: server,
		Service: service,

	}
	h.registerRoutes()

	return h
}


func (h *UserHandler) registerRoutes() {
	h.Server.HandleFunc("POST /users", h.CreateUser)
	h.Server.HandleFunc("GET /users/{id}", h.GetUser)
	h.Server.HandleFunc("PUT /users/{id}", h.UpdateUser)
	h.Server.HandleFunc("DELETE /users/{id}", h.DeleteUser)
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

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println("User id",id)

	w.WriteHeader(http.StatusNoContent) 
}