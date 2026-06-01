// eventstream/kafka.go
package eventstream

import (
	"context"
	"fmt"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaClient struct {
	client  *kgo.Client
	brokers []string
}

func NewKafkaClient(brokers []string) (*KafkaClient, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init kafka client: %w", err)
	}

	return &KafkaClient{
		client:  cl,
		brokers: brokers,
	}, nil
}

func (k *KafkaClient) Close() {
	k.client.Close()
}

func (k *KafkaClient) AddEvent(ctx context.Context, topic string, key []byte, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	results := k.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("failed to produce message to %s: %w", topic, err)
	}

	return nil
}



func (k *KafkaClient) ConsumeEvents(ctx context.Context, topics []string, groupID string, handler func(topic string, key []byte, value []byte)) error {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(k.brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
	)
	if err != nil {
		return fmt.Errorf("failed to init kafka consumer group: %w", err)
	}
	defer cl.Close()

	log.Printf("Kafka consumer started on topics: %v under group: %s", topics, groupID)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := cl.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				log.Printf("Kafka fetch errors: %v", errs)
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				handler(record.Topic, record.Key, record.Value)
			}
			if err := cl.CommitUncommittedOffsets(ctx); err != nil {
				log.Printf("failed to commit offsets: %v", err)
			}
		}
	}
}