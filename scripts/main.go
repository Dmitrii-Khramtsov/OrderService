// github.com/Dmitrii-Khramtsov/orderservice/scripts/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}


func generateRandomPhone() string {
	const digits = "0123456789"
	phone := make([]byte, 10)
	for i := range phone {
		phone[i] = digits[rand.Intn(len(digits))]
	}
	return "+" + string(phone)
}

func generateRandomOrderItem() entities.Item {
	return entities.Item{
		ChrtID:      rand.Intn(10000000),
		Name:        "Product_" + generateRandomString(5),
		TrackNumber: "TRACK_" + generateRandomString(10),
		Price:       rand.Intn(1000) + 100,
		RID:         uuid.New().String(),
		Sale:        rand.Intn(50),
		Size:        generateRandomString(3),
		TotalPrice:  rand.Intn(1000) + 100,
		NmID:        rand.Intn(1000000),
		Brand:       "Brand_" + generateRandomString(5),
		Status:      rand.Intn(5),
	}
}

func generateRandomOrder() entities.Order {
	return entities.Order{
		OrderUID:        uuid.New().String(),
		TrackNumber:     "WBILMTESTTRACK_" + generateRandomString(5),
		Entry:           uuid.New().String(),
		Locale:          "en",
		InternalSig:     uuid.New().String(),
		CustomerID:      "customer_" + generateRandomString(8),
		DeliveryService: "delivery_service_" + generateRandomString(5),
		ShardKey:        generateRandomString(5),
		SMID:            rand.Intn(100),
		DateCreated:     time.Now().Format(time.RFC3339),
		OOFShard:        generateRandomString(5),
		Delivery: entities.Delivery{
			Name:    "Customer " + generateRandomString(5),
			Phone:   generateRandomPhone(),
			Zip:     generateRandomString(6),
			City:    "City_" + generateRandomString(5),
			Address: "Address_" + generateRandomString(10),
			Region:  "Region_" + generateRandomString(5),
			Email:   generateRandomString(5) + "@example.com",
		},
		Payment: entities.Payment{
			Transaction:  uuid.New().String(),
			RequestID:    uuid.New().String(),
			Currency:     "USD",
			Provider:     "payment_provider_" + generateRandomString(5),
			Amount:       rand.Intn(1000) + 100,
			PaymentDT:    time.Now().Unix(),
			Bank:         "Bank_" + generateRandomString(5),
			DeliveryCost: rand.Intn(100),
			GoodsTotal:   rand.Intn(1000) + 100,
			CustomFee:    rand.Intn(50),
		},
		Items: []entities.Item{
			generateRandomOrderItem(),
			generateRandomOrderItem(),
		},
	}
}

func sendOrderToKafka(writer *kafka.Writer, order entities.Order) error {
	messageValue, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(order.OrderUID),
		Value: messageValue,
	})
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	brokerAddress := os.Getenv("KAFKA_BROKERS")
	if brokerAddress == "" {
		log.Fatal("Kafka broker address is not set in .env")
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "orders"
	}

	numMessages, _ := strconv.Atoi(os.Getenv("NUMBER_OF_MESSAGES"))
	if numMessages <= 0 {
		numMessages = 10
	}

	delayMs, _ := strconv.Atoi(os.Getenv("DELAY_MS"))
	if delayMs <= 0 {
		delayMs = 500
	}

	log.Printf("Using Kafka broker: %s", brokerAddress)
	log.Printf("Using Kafka topic: %s", topic)

	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Fatalf("Failed to get controller: %v", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Fatalf("Failed to connect to controller: %v", err)
	}
	defer controllerConn.Close()

	_ = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokerAddress},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	for i := 0; i < numMessages; i++ {
		order := generateRandomOrder()
		if err := sendOrderToKafka(writer, order); err != nil {
			log.Printf("Failed to send order %s: %v", order.OrderUID, err)
			continue
		}
		log.Printf("Successfully sent order: %s", order.OrderUID)
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
}