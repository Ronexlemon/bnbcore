package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)

type RegisterTenantRequest struct {
	ShopName   string `json:"shop_name"`
	Subdomain  string `json:"subdomain"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type TenantHandler struct{
	Server *http.ServeMux
	Service *tenant.Service
	JWTAuthManager *auth.JwtManager
}

func NewTenantHandler(server *http.ServeMux,service *tenant.Service,m *auth.JwtManager)(*TenantHandler){
	h:=&TenantHandler{
		Server: server,
		Service: service,
		JWTAuthManager: m,
	}
	h.registerHandler()

	return  h
}

func(h *TenantHandler) registerHandler(){
	api := "/api/v1"
	h.Server.HandleFunc(api+"/tenant/register", h.RegisterTenantWithUser)
}

func (h *TenantHandler)  RegisterTenantWithUser(w http.ResponseWriter,r *http.Request){

	if r.Method != http.MethodPost{
		http.Error(w,"method not allowed",http.StatusMethodNotAllowed)
		return
	}

	var req RegisterTenantRequest

	err:=json.NewDecoder(r.Body).Decode(&req)
	if err !=nil{
		http.Error(w,"invalid request body",http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	err = h.Service.RegisterTenantWithUser(
		ctx,
		req.ShopName,
		req.Subdomain,
		req.Email,
		req.Password,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "tenant created successfully",
	})


}