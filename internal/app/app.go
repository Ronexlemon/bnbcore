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
	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/domain/services"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/domain/unit"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/handler"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/repository"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/upload"
	"github.com/ronexlemon/bnbcore/internal/senders"
	"github.com/ronexlemon/bnbcore/internal/worker"
)

func NewMuxService(ctx context.Context) http.Handler { 
	workerCtx := context.Background()
    config_env := config.Load()
	rpConfig := config.LoadKafkaConfig()
	sender :=senders.NewSender(senders.Config{Host: config_env.HOST,Port: config_env.PORT,From: config_env.FROM,Password: config_env.PASSWORD,Username: config_env.USERNAME})
	stream, err := eventstream.NewKafkaClient(rpConfig.Brokers)
    if err != nil {
        log.Fatalf("failed to init event stream: %v", err)
    }
	redis_client :=config.NewRedisClient(config_env.REDIS_URL)

	media,err := upload.NewMediaService(config_env.CLOUDINARY_URL,redis_client.Client)
	if err != nil {
        log.Fatalf("failed to init media : %v", err)
    }

    var peppers map[string]string
    err = json.Unmarshal([]byte(config_env.PASSWORD_PEPPERS), &peppers)
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

	unit_service_repo, err:= repository.NewUnitServiceRepository(conn)
    if err != nil {
        log.Fatalf("Failed to initialize unit service repository: %v", err)
    }
	subscription_repo, err:= repository.NewSubscriptionRepository(conn)
    if err != nil {
        log.Fatalf("Failed to initialize  Subscription repository: %v", err)
    }
	notification_repo, err:= repository.NewNotificationRepository(conn)
    if err != nil {
        log.Fatalf("Failed to initialize  notification repository: %v", err)
    }

    passwordEngine, err := password.NewPasswordHasher(peppers, config_env.ACTIVE_PEPPER_VERSION)
    if err != nil {
        log.Fatalf("password pepper and active version needed: %v", err)
    }

    tokenEngine, err := token.NewTokenHasher(config_env.MasterKeyHex, config_env.EncryptedDataKeyHex)
    if err != nil {
        log.Fatalf("master key and active encryptedHex need needed: %v", err)
    }

    tenant_service := tenant.NewService(tenant_repo)
	 sunscription_service := subscription.NewService(subscription_repo)
	 notification_service := notification.NewService(notification_repo)
	unit_service := unit.NewUnitService(unit_repo)
	unit_service_service := services.NewService(unit_service_repo)
	booking_service := booking.NewBookingService(booking_repo)
    user_service := user.NewUserservice(user_repo, passwordEngine, tokenEngine, config_env.GOOGLE_CLIENT_ID)

     _ = handler.NewTenantHandler(mux, tenant_service, jwtManager,subscription_repo,stream)
     _ = handler.NewUserHandler(mux, user_service, jwtManager,config_env.BASE_DOMAIN,stream,sender)
	 _ = handler.NewUnitHandler(mux, unit_service, jwtManager,subscription_repo,stream,media)
	 _ = handler.NewBookingHandler(mux, booking_service, jwtManager,stream,subscription_repo)
	 _ = handler.NewRoomServiceHandler(mux,unit_service_service, jwtManager,subscription_repo,stream)
	 _ = handler.NewSubscriptionHandler(mux,sunscription_service, jwtManager,stream)
	 _ = handler.NewNotificationHandler(mux,notification_service, jwtManager)

	  waWorker := worker.NewBookingNotificationWorker(stream, worker.WhatsAppConfig{
        AccountSID: config_env.TWILIO_ACCOUNT_SID,
        AuthToken:  config_env.TWILIO_AUTH_TOKEN,
        FromNumber: config_env.TWILIO_WHATSAPP_FROM,
    },notification_service)
	subWorker := worker.NewSubscriptionExpiryWorker(subscription_repo,time.Duration(time.Second *30))
	generalNotificationWorker := worker.NewNotificationWorker(stream,notification_service,sender)
    go func() {
        if err := waWorker.Start(workerCtx); err != nil {
            log.Printf("booking notification worker stopped: %v", err)
        }
    }()
	 go func() {
        subWorker.Start(workerCtx)
    }()
	go func() {
       
		 if err :=  generalNotificationWorker.Start(workerCtx); err != nil {
            log.Printf("general notification worker stopped: %v", err)
        }
    }()

    return auth.SubdomainResolver(tenant_service, config_env.BASE_DOMAIN)(mux)
}