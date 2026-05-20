package app

import (
	"context"
	"log"
	"net/http"

	"github.com/ronexlemon/bnbcore/internal/config"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/handler"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/repository"
)


func NewMuxService(ctx context.Context)*http.ServeMux{
	
   config_env:= config.Load()
	conn,err:=db.NewPostgressConnect(ctx,config_env)
	if err !=nil{
		log.Fatalf("Failed to connect to db: %v", err)
	}
	
	mux:=http.NewServeMux()
    tenant_repo,err:= repository.NewTenantRepository(conn)
	if err != nil {
		log.Fatalf("Failed to initialize tenant repository: %v", err)
	}
	tenant_service := tenant.NewService(tenant_repo)
	_ = handler.NewTenantHandler(mux, tenant_service)




	return mux

}