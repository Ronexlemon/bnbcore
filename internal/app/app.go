package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/config"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/handler"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/repository"
)


func NewMuxService(ctx context.Context)*http.ServeMux{
	
   config_env:= config.Load()
   
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
	tenant_service := tenant.NewService(tenant_repo)
	user_service := user.NewUserservice(user_repo)
	_ = handler.NewTenantHandler(mux, tenant_service,jwtManager)
	_ =handler.NewUserHandler(mux,user_service,jwtManager)




	return mux

}