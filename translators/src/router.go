package src

import (
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
	"translators/src/di"
	"translators/src/dto"
	"translators/src/services"
	"translators/src/utils"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Handler interface {
	GetMethods() []dto.Routes
}

func NewRouter(out chan string) *gin.Engine {
	router := gin.Default()
	router.Use(gin.Recovery())

	logger := utils.ProvideLogger()
	container := di.Container{Log: logger}

	producer := container.GetKafkaProducer()
	consumer := container.GetKafkaConsumer()

	go func() {
		defer producer.Close()
		produceToKafka(producer, logger)
		services.HandleKafkaEvents(producer, logger)
	}()

	go func() {
		defer consumer.Close()
		consumeFromKafkaToChan(consumer, out, logger)
	}()

	go func(out chan string) {
		ln, err := net.Listen("tcp", ":"+os.Getenv("TRANSLATOR_TCP_PORT"))
		if err != nil {
			logger.Error("Can't create net.Listener", zap.Error(err))
			os.Exit(1)
			return
		}

		defer ln.Close()

		for {
			conn, err := ln.Accept()
			if err != nil {
				logger.Error("Can't accept connection", zap.Error(err))
			}
			go func(conn *net.Conn, data chan string) {
				defer (*conn).Close()
				for v := range data {
					(*conn).Write([]byte(v))
				}
			}(&conn, out)
		}
		// TODO: send from this retranslator to consumers
	}(out)

	subscribe_handler := container.GetSubscribeHandler()

	for _, route := range GetRoutes(subscribe_handler) {
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			router.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Pattern, route.HandlerFunc)
		}
	}
	router.GET("/healthcheck", Healthcheck)
	return router
}

func GetRoutes(handlers ...Handler) []dto.Routes {
	allRoutes := []dto.Routes{}
	for _, handler := range handlers {
		allRoutes = append(allRoutes, (handler.GetMethods())...)
	}
	return allRoutes
}

func Healthcheck(c *gin.Context) {
	c.String(http.StatusOK, "alive")
}

func consumeFromKafkaToChan(consumer *kafka.Consumer, out chan string, logger *zap.Logger) {
	for {
		msg, err := consumer.ReadMessage(100 * time.Millisecond)
		if err != nil {
			logger.Error("Error consuming message", zap.Error(err))
			continue
		}
		out <- string(msg.Value)
	}
}

func produceToKafka(producer *kafka.Producer, logger *zap.Logger) {
	topic := "transactions"
	deliveryChan := make(chan kafka.Event, 10000)
	for {
		producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          RandStringBytesRmndr(10000),
		}, deliveryChan)
	}
}

func RandStringBytesRmndr(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return b
}
