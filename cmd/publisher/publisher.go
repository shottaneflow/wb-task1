package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"task1/internal/order"
	"time"

	"task1/internal/config"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

var orderChannel = make(chan order.Order, 100)
var errChan = make(chan error, 2)

func main() {

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	time.Sleep(20 * time.Second)
	logger := slog.Default()
	err := godotenv.Load()
	if err != nil {
		logger.Error("Ошибка загрузки переменных окружение", "error", err.Error())
		errChan <- err
	}
	cfg, err := config.MustLoad()
	if err != nil {
		logger.Error("Ошибка при загрузке конфига", "error", err)
		errChan <- err
	}
	ctx := context.Background()
	gracefulShutdown := func() {
		close(orderChannel)
		close(errChan)
	}
	writer := createKafkaWriter()
	defer writer.Close()
	go sendMessageToKafka(ctx, cfg.TimeDurationPublisher, writer, logger)
	select {
	case <-osSignal:
		gracefulShutdown()
		os.Exit(0)
	case <-errChan:
		gracefulShutdown()
		os.Exit(1)
	}

}
func generateRandomOrder() order.Order {
	trackNumber := fmt.Sprintf("WBIL%d", rand.Intn(10000))
	return order.Order{
		TrackNumber:       trackNumber,
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: fmt.Sprintf("SIG%d", rand.Intn(1000)),
		CustomerID:        fmt.Sprintf("customer%d", rand.Intn(1000)),
		DeliveryService:   "meest",
		ShardKey:          fmt.Sprintf("%d", rand.Intn(10)),
		SmID:              rand.Intn(100),
		DateCreated:       time.Now(),
		OofShard:          fmt.Sprintf("%d", rand.Intn(2)),
		Delivery: &order.Delivery{
			Name:    fmt.Sprintf("Customer %d", rand.Intn(1000)),
			Phone:   fmt.Sprintf("+7%d", 900000000+rand.Intn(10000000)),
			Zip:     fmt.Sprintf("%d", 100000+rand.Intn(900000)),
			City:    []string{"Moscow", "SPb", "Kazan", "Novosibirsk"}[rand.Intn(4)],
			Address: fmt.Sprintf("Street %d, building %d", rand.Intn(100), rand.Intn(50)),
			Region:  "Russia",
			Email:   fmt.Sprintf("customer%d@example.com", rand.Intn(1000)),
		},
		Payment: &order.Payment{
			Transaction:  uuid.New().String(),
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       rand.Intn(10000) + 100,
			PaymentDT:    int(time.Now().Unix()),
			Bank:         []string{"alpha", "sber", "tinkoff", "vtb"}[rand.Intn(4)],
			DeliveryCost: rand.Intn(500) + 100,
			GoodsTotal:   rand.Intn(500) + 50,
			CustomFee:    rand.Intn(50),
		},
		Items: generateRandomItems(),
	}
}
func generateRandomItems() []*order.Item {
	numItems := rand.Intn(3) + 1
	items := make([]*order.Item, numItems)

	for i := 0; i < numItems; i++ {
		items[i] = &order.Item{
			ChrtID:      rand.Intn(10000000),
			TrackNumber: fmt.Sprintf("WBIL%d", rand.Intn(10000)),
			Price:       rand.Intn(1000) + 10,
			Rid:         uuid.New().String(),
			Name:        []string{"Phone", "Laptop", "Book", "Clothes", "Shoes"}[rand.Intn(5)],
			Sale:        rand.Intn(50),
			Size:        fmt.Sprintf("%d", rand.Intn(5)),
			TotalPrice:  rand.Intn(1000) + 50,
			NmID:        rand.Intn(1000000),
			Brand:       []string{"Apple", "Samsung", "Nike", "Adidas", "Sony"}[rand.Intn(5)],
			Status:      202,
		}
	}
	return items
}

func createKafkaWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "my-topic",
	})
}

func sendMessageToKafka(ctx context.Context, interavl time.Duration, writer *kafka.Writer, logger *slog.Logger) {
	ticker := time.NewTicker(interavl)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ord := generateRandomOrder()
			orderBytes, err := json.Marshal(ord)
			if err != nil {
				logger.Error("Ошибка маршалинга order", "error", err)
				errChan <- err
			}
			err = writer.WriteMessages(ctx, kafka.Message{
				Value: orderBytes,
			})
			if err != nil {
				logger.Error("Ошибка при отправке:", "error", err)
				errChan <- err
			}
			logger.Info("Заказ отправлен")

		}
	}
}
