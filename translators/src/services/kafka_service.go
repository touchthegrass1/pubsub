package services

import (
	"encoding/json"
	"time"
	"translators/src/dto"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type KafkaTransactionService struct {
	producer     *kafka.Producer
	consumer     *kafka.Consumer
	deliveryChan chan kafka.Event
	log          *zap.Logger
}

type KafkaSender interface {
	SendMessage(dto.Subscription)
}

type KafkaReader interface {
	ReadMessage(string)
}

func (service KafkaTransactionService) SendMessage(message dto.Subscription) error {
	topic := "transactions"
	value, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = service.producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          value,
		},
		service.deliveryChan,
	)
	if err != nil {
		service.log.Error("Error sending kafka message", zap.Error(err))
		return err
	}
	return nil
}

func (service KafkaTransactionService) ReadMessage(topic string) (*kafka.Message, error) {
	ev, err := service.consumer.ReadMessage(100 * time.Millisecond)
	if err != nil {
		service.log.Error("Error with event", zap.Error(err))
		return nil, err
	}

	return ev, nil
}

func HandleKafkaEvents(producer *kafka.Producer, logger *zap.Logger) {
	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				logger.Error("Failed to deliver message: ", zap.Error(ev.TopicPartition.Error))
			} else {
				logger.Info("Produced event",
					zap.String("topic", *ev.TopicPartition.Topic),
					zap.String("key", string(ev.Key)),
					zap.String("value", string(ev.Value)),
				)
			}
		}
	}
}

func ProvideKafkaTransactionService(producer *kafka.Producer, consumer *kafka.Consumer, delChan chan kafka.Event, logger *zap.Logger) KafkaTransactionService {
	return KafkaTransactionService{producer: producer, consumer: consumer, deliveryChan: delChan, log: logger}
}
