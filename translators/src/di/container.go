package di

import (
	"bufio"
	"os"
	"strings"

	"translators/src/dblayer"
	"translators/src/handlers"
	"translators/src/repositories"
	"translators/src/services"
	"translators/src/utils"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
	db       *gorm.DB
	Log      *zap.Logger
}

func (container Container) GetSubscribeHandler() handlers.SubscribeHandler {
	logger := utils.ProvideLogger()
	repo := container.GetSubscribeRepository()
	return handlers.SubscribeHandler{Repo: repo, Log: logger}
}

func (container Container) GetSubscribeRepository() repositories.SubscribeRepository {
	db := container.GetDB()
	logger := utils.ProvideLogger()
	return repositories.SubscribeRepository{Db: db, Log: logger}
}

func (container Container) GetKafkaTransactionService() services.KafkaTransactionService {
	consumer := container.GetKafkaConsumer()
	producer := container.GetKafkaProducer()
	deliveryChan := make(chan kafka.Event, 10000)
	logger := utils.ProvideLogger()
	return services.ProvideKafkaTransactionService(producer, consumer, deliveryChan, logger)
}

func (container *Container) GetKafkaProducer() *kafka.Producer {
	if container.producer != nil {
		return container.producer
	}
	config := container.GetKafkaConfig()
	producer, err := kafka.NewProducer(&config)
	if err != nil {
		panic(err)
	}
	container.producer = producer
	go services.HandleKafkaEvents(producer, container.Log)
	return producer
}

func (container *Container) GetDB() *gorm.DB {
	if container.db != nil {
		return container.db
	}
	logger := utils.ProvideLogger()
	dbparams := dblayer.ProvideDBParams()
	gormConfig := dblayer.ProvideGormConfig()
	database, err := dblayer.ProvideDB(dbparams, logger, gormConfig)
	if err != nil {
		panic(err)
	}
	container.db = database
	return database
}

func (container *Container) GetKafkaConsumer() *kafka.Consumer {
	if container.consumer != nil {
		return container.consumer
	}
	config := container.GetKafkaConfig()
	config["auto.offset.reset"] = "earliest"
	config["group.id"] = "test-consumer-group"

	consumer, err := kafka.NewConsumer(&config)
	if err != nil {
		panic(err)
	}
	topic := "transactions"
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		panic(err)
	}
	container.consumer = consumer
	return consumer
}

func (container Container) GetKafkaConfig() kafka.ConfigMap {
	m := make(map[string]kafka.ConfigValue)

	file, err := os.Open("kafka.properties")
	if err != nil {
		container.Log.Error("Error getting kafka config", zap.Error(err))
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}
		kv := strings.Split(line, "=")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		m[key] = value
	}
	return m
}
