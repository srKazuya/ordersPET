package orderSaver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"
	"github.com/srKazuya/ordersPET/internal/storage"
)

type Saver struct {
	log     *slog.Logger
	storage OrderSaver
}

type OrderSaver interface {
	SaveOrder(ctx context.Context, order *storage.Order) error
}

func New(log *slog.Logger, saver OrderSaver) *Saver {
	return &Saver{
		log:     log,
		storage: saver,
	}
}

func (s *Saver) SaveOrder(msg []byte, offset *kafka.Offset) error {
	const op = "orderSaver.GetOrder"

	
	var order storage.Order
	if err := json.Unmarshal(msg, &order); err != nil {
		s.log.Error("failed to unmarshal kafka msg", sl.Err(err))
		return fmt.Errorf("%s: failed to unmarshal message: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.storage.SaveOrder(ctx, &order); err != nil {
		s.log.Error("failed to save order", sl.Err(err))
		return fmt.Errorf("%s: failed to save order: %w", op, err)
	}

	s.log.Info("order saver successfully", slog.String("order_id", order.OrderUID))
	return nil

}

