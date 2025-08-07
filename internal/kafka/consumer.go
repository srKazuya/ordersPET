package kafka

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"
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
		"auto.offset.reset": "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		log.Error("failed to create new consumer", sl.Err(err))
		return nil, fmt.Errorf("%s error with new Consumer: %w", op, err)
	}

	if err = c.Subscribe(topic, nil); err != nil {
		log.Error("failed to subscribe on topic", sl.Err(err))
		return nil, fmt.Errorf("%s failed to subscribe on topic: %s with new Consumer: %w", op, topic, err)
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
			log.Error("failed to read msg", sl.Err(err))
		}
		if kafkaMsg == nil {
			continue
		}
		if err := c.Service.SaveOrder(kafkaMsg.Value, &kafkaMsg.TopicPartition.Offset); err != nil {
			log.Error("failed to save order", sl.Err(err))
			continue
		}
		if _, err = c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Error("failed to save offset", sl.Err(err))
			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}
	return c.consumer.Close()
}
