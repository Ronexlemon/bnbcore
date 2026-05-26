package eventstream

import (
	"context"
	"encoding/json"
	"fmt"

	
)

func (k *KafkaClient) Publish(ctx context.Context, topic string, key string, payload any) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return k.AddEvent(ctx, topic, []byte(key), value)
}