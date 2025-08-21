package cache

import (
	"context"
	"log/slog"
	"sync"
	"task1/internal/order"
)

type OrderCache struct {
	store  map[string]order.Order
	rw     *sync.RWMutex
	logger *slog.Logger
}

func NewOrderCache(logger *slog.Logger) *OrderCache {
	return &OrderCache{store: make(map[string]order.Order), rw: new(sync.RWMutex), logger: logger}
}

func (cache *OrderCache) Store(order order.Order) {
	cache.rw.Lock()
	cache.store[order.OrderUID] = order
	cache.rw.Unlock()
	cache.logger.Info("Положили order в кеш")
}

func (cache *OrderCache) Load(id string) (order.Order, bool) {
	cache.rw.RLock()
	order, ok := cache.store[id]
	if ok {
		cache.logger.Info("Взяли order из кеша")
	}
	return order, ok
}

func (cache *OrderCache) RestoreFromDB(ctx context.Context, repos order.Repository) {
	var orders []order.Order
	orders, err := repos.FindAll(ctx)
	if err != nil {
		cache.logger.Error("Ошибка выгрузки orders из бд в кеш", "error", err)
		return
	}
	for _, order := range orders {
		cache.rw.Lock()
		cache.store[order.OrderUID] = order
		cache.rw.Unlock()
	}
	cache.logger.Info("Загрузили orders from db")
}
