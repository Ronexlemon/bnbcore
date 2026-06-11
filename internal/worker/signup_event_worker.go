
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/queue"
	
)



type SignupRegistrationWorker struct {
	Stream   *eventstream.KafkaClient
	Enqueuer  *queue.Enqueuer
}

func NewSignupRegistrationWorker(stream *eventstream.KafkaClient,notification_service *notification.Service,enqueuer *queue.Enqueuer) *SignupRegistrationWorker{
	return &SignupRegistrationWorker{
		Stream:   stream,
		Enqueuer: enqueuer,
	}
}

func (w *SignupRegistrationWorker) Start(ctx context.Context) error {
	topics := []string{
		eventstream.TopicUserSignedUp,
	}

	return w.Stream.ConsumeEvents(ctx, topics, "signup-registration-group", func(topic string, key, value []byte) {
		switch topic {
		case eventstream.TopicUserSignedUp:
			var event eventstream.UserSignedUpEvent
			if err := json.Unmarshal(value, &event); err != nil {
				log.Printf("failed to unmarshal signup created event: %v", err)
				return
			}
			fmt.Println("EVENT  for signup INCOMING", event)
			w.enqueueSignupRegistration(ctx, event)

		default:
			log.Printf("unhandled topic: %s", topic)
		}
	})
}


//new
func (w *SignupRegistrationWorker) enqueueSignupRegistration(ctx context.Context, event eventstream.UserSignedUpEvent) {
	if event.Email == "" {
		log.Printf("[consumer] no Email for signup %s, skipping", event.Email)
		return
	}

	

	payload := UserSignupPayload{
		UserID: event.UserID,
		Email: event.Email,
		Link: event.SignupLink,
	}

	if _, err := queue.EnqueueTask(ctx, w.Enqueuer,UserSignUpTask, payload); err != nil {
		log.Printf("[consumer] failed to enqueue created task for signup %s: %v", event.UserID, err)
		return
	}

	log.Printf("[consumer] job enqueued for booking %s", event.UserID)
}

