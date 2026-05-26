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
	"github.com/ronexlemon/bnbcore/internal/domain/booking"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/domain/unit"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/handler"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/repository"
)

func NewMuxService(ctx context.Context) http.Handler { 
    config_env := config.Load()

    var peppers map[string]string
    err := json.Unmarshal([]byte(config_env.PASSWORD_PEPPERS), &peppers)
    if err != nil {
        log.Fatalf("Critical error parsing PASSWORD_PEPPERS JSON structure: %v", err)
    }

    jwtManager := auth.NewJwtManager(
        []byte(config_env.JWT_SECRET),
        time.Duration(config_env.JWT_ACCESS_DURATION_MINUTE)*time.Minute,
        time.Duration(config_env.JWT_REFRESH_DURATION_HOUR)*24*time.Hour,
    )

    conn, err := db.NewPostgressConnect(ctx, config_env)
    if err != nil {
        log.Fatalf("Failed to connect to db: %v", err)
    }

    mux := http.NewServeMux()

    tenant_repo, err := repository.NewTenantRepository(conn)
    if err != nil {
        log.Fatalf("Failed to initialize tenant repository: %v", err)
    }

    user_repo, err := repository.NewUserRepository(conn)
    if err != nil {
        log.Fatalf("Failed to initialize user repository: %v", err)
    }
	unit_repo, err:= repository.NewUnitRepository(conn)
    if err != nil {
        log.Fatalf("Failed to initialize unit repository: %v", err)
    }
	booking_repo, err:= repository.NewBookingRepository(conn)
    if err != nil {
        log.Fatalf("Failed to initialize booking repository: %v", err)
    }

    passwordEngine, err := password.NewPasswordHasher(peppers, config_env.ACTIVE_PEPPER_VERSION)
    if err != nil {
        log.Fatalf("password pepper and active version needed: %v", err)
    }

    tokenEngine, err := token.NewTokenHasher(config_env.MasterKeyHex, config_env.EncryptedDataKeyHex)
    if err != nil {
        log.Fatalf("master key and active encryptedHex need needed: %v", err)
    }

    tenant_service := tenant.NewService(tenant_repo, passwordEngine)
	unit_service := unit.NewUnitService(unit_repo)
	booking_service := booking.NewBookingService(booking_repo)
    user_service := user.NewUserservice(user_repo, passwordEngine, tokenEngine, config_env.GOOGLE_CLIENT_ID)

    _ = handler.NewTenantHandler(mux, tenant_service, jwtManager)
    _ = handler.NewUserHandler(mux, user_service, jwtManager,config_env.BASE_DOMAIN)
	 _ = handler.NewUnitHandler(mux, unit_service, jwtManager)
	  _ = handler.NewBookingHandler(mux, booking_service, jwtManager)

    return auth.SubdomainResolver(tenant_service, config_env.BASE_DOMAIN)(mux)
}