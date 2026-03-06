package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Producer wraps Sarama SyncProducer
type Producer struct {
	producer sarama.SyncProducer
	brokers  []string
}

// Config holds Kafka configuration
type Config struct {
	Brokers []string
	Enabled bool
}

var DefaultProducer *Producer

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	p := &Producer{
		producer: producer,
		brokers:  brokers,
	}

	log.Printf("Connected to Kafka brokers: %v", brokers)
	return p, nil
}

// Connect initializes the default Kafka producer
func Connect(cfg *Config) error {
	if !cfg.Enabled || len(cfg.Brokers) == 0 {
		log.Printf("Kafka disabled or no brokers configured")
		return nil
	}

	producer, err := NewProducer(cfg.Brokers)
	if err != nil {
		return err
	}

	DefaultProducer = producer
	return nil
}

// SendEvent sends a message to Kafka topic (implements KafkaClient interface)
func (p *Producer) SendEvent(topic string, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("[KAFKA ERROR] Failed to send message to %s: %v", topic, err)
		return err
	}

	log.Printf("[KAFKA] Sent to %s (partition=%d, offset=%d)", topic, partition, offset)
	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// CloseDefault closes the default producer
func CloseDefault() error {
	if DefaultProducer != nil {
		return DefaultProducer.Close()
	}
	return nil
}

// IsConnected checks if Kafka is connected
func IsConnected() bool {
	return DefaultProducer != nil
}

// GetProducer returns the default producer (implements KafkaClient interface)
func GetProducer() *Producer {
	return DefaultProducer
}
