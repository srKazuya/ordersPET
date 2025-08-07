package kafka

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"
)

const flushTimeout = 5000

var errUnknownType = errors.New("unkniwn event type")

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(log *slog.Logger, address []string) (*Producer, error) {
	const op = "kafka.producer"

	log = log.With(
		slog.String("op", op),
	)

	cfg := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(address, ","),
	}

	p, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Error("failed to create new producer", sl.Err(err))
		return nil, fmt.Errorf("%s error with new Producer: %w", op, err)
	}

	return &Producer{producer: p}, nil
}

func (p *Producer) Produce(message, topic string) error {
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(message),
		Key:   nil,
	}
	kafkaChan := make(chan kafka.Event)
	if err := p.producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return err
	}
	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return ev
	default:
		return errUnknownType
	}
}

func (p *Producer) Close(){
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}