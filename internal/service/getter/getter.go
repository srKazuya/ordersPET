package orderGetter

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"sync"

	
	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"
	"github.com/srKazuya/ordersPET/internal/storage"
)

type Getter struct {
	log     *slog.Logger
	storage OrderGetter
	
	cache struct {
		sync.RWMutex
		data map[string]*storage.Order
	}
}

type OrderGetter interface {
	GetOrderByUID(ctx context.Context, orderUID string) (storage.Order, error)
	
}

func New(log *slog.Logger, getter OrderGetter) *Getter {
	g := &Getter{
		log:     log,
		storage: getter,
	}
	g.cache.data = make(map[string]*storage.Order)
	return g
}

func (g *Getter) GetOrderByUID(ctx context.Context, orderUID string) (storage.Order, error) {
	const op = "orderGetter.GetOrderByUID"
	fmt.Println("GetOrderByUID вызван")
	
	g.cache.RLock()
	order, found := g.cache.data[orderUID]
	g.cache.RUnlock()
	if found {
		g.log.Info("order found in cache", slog.String("order_id", orderUID))
		return *order, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	orderVal, err := g.storage.GetOrderByUID(ctx, orderUID)
	if err != nil {
		g.log.Error("failed to get order", sl.Err(err))
		return storage.Order{}, fmt.Errorf("%s: failed to get order: %w", op, err)
	}

	g.log.Info("order retrieved from storage", slog.String("order_id", orderVal.OrderUID))

	g.cache.Lock()
	g.cache.data[orderVal.OrderUID] = &orderVal
	g.cache.Unlock()

	g.log.Info("order cached successfully")

	return orderVal, nil
}

