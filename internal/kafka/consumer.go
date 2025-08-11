package kafka

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"
)

var (
	ErrCreateConsumer = errors.New("failed to create Kafka consumer")
	ErrSubscribeTopic = errors.New("failed to subscribe on Kafka topic")
	ErrReadMessage    = errors.New("failed to read Kafka message")
	ErrSaveOrder      = errors.New("failed to save order")
	ErrSaveOffset     = errors.New("failed to store offset")
)

const (
	sessionTimeout = 7000
	noTimeout      = -1
)

type Consumer struct {
	consumer *kafka.Consumer
	Service  OrderSaver
	stop     bool
}

type OrderSaver interface {
	SaveOrder(message []byte, offset *kafka.Offset) error
}

func NewConsumer(saver OrderSaver, log *slog.Logger, address []string, topic, consumerGroup string) (*Consumer, error) {
	const op = "kafka.consumer"

	log = log.With(
		slog.String("op", op),
	)

	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(address, ","),
		"group.id":                 consumerGroup,
		"session.timeout.ms":       sessionTimeout,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		log.Error("failed to create new consumer", sl.Err(err))
		return nil, fmt.Errorf("%s: %w: %v", op, ErrCreateConsumer, err)
	}

	if err = c.Subscribe(topic, nil); err != nil {
		log.Error("failed to subscribe on topic", sl.Err(err))
		return nil, fmt.Errorf("%s: %w: topic=%s: %v", op, ErrSubscribeTopic, topic, err)
	}

	return &Consumer{
		consumer: c,
		Service:  saver,
	}, nil
}

func (c *Consumer) Start(log *slog.Logger) {
	for {
		if c.stop {
			break
		}

		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			log.Error("read message error", sl.Err(err))
			err = fmt.Errorf("%w: %v", ErrReadMessage, err)
			continue
		}

		if kafkaMsg == nil {
			continue
		}

		if err := c.Service.SaveOrder(kafkaMsg.Value, &kafkaMsg.TopicPartition.Offset); err != nil {
			log.Error("save order error", sl.Err(err))
			err = fmt.Errorf("%w: %v", ErrSaveOrder, err)
			continue
		}

		if _, err = c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Error("save offset error", sl.Err(err))
			err = fmt.Errorf("%w: %v", ErrSaveOffset, err)
			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}
	return c.consumer.Close()
}
