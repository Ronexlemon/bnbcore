package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/auth/password"
	"github.com/ronexlemon/bnbcore/internal/auth/token"
	"github.com/ronexlemon/bnbcore/internal/config"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/handler"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/repository"
)


func NewMuxService(ctx context.Context)*http.ServeMux{
	
   config_env:= config.Load()
   var peppers map[string]string
	err := json.Unmarshal([]byte(config_env.PASSWORD_PEPPERS), &peppers)
	if err != nil {
		log.Fatalf("Critical error parsing PASSWORD_PEPPERS JSON structure: %v", err)
	}
   jwtManager := auth.NewJwtManager(
		[]byte(config_env.JWT_SECRET),
		time.Duration(config_env.JWT_ACCESS_DURATION_MINUTE)*time.Minute,   // access token TTL
		time.Duration(config_env.JWT_REFRESH_DURATION_HOUR)*24*time.Hour,   // refresh token TTL
	)

	conn,err:=db.NewPostgressConnect(ctx,config_env)
	if err !=nil{
		log.Fatalf("Failed to connect to db: %v", err)
	}
	
	mux:=http.NewServeMux()
    tenant_repo,err:= repository.NewTenantRepository(conn)

	if err != nil {
		log.Fatalf("Failed to initialize tenant repository: %v", err)
	}
	user_repo,err :=repository.NewUserRepository(conn)
	if err != nil {
		log.Fatalf("Failed to initialize user repository: %v", err)
	}
	passwordEngine,err :=password.NewPasswordHasher(peppers,config_env.ACTIVE_PEPPER_VERSION)
	if err != nil {
		log.Fatalf("password pepper and active version needed: %v", err)
	}
	 tokenEngine,err:=token.NewTokenHasher(config_env.MasterKeyHex,config_env.EncryptedDataKeyHex)
	 if err != nil {
		log.Fatalf("master key and active encryptedHex need needed: %v", err)
	}
	tenant_service := tenant.NewService(tenant_repo,passwordEngine)
	user_service := user.NewUserservice(user_repo,passwordEngine,tokenEngine)
	
	_ = handler.NewTenantHandler(mux, tenant_service,jwtManager)
	_ =handler.NewUserHandler(mux,user_service,jwtManager)




	return mux

}