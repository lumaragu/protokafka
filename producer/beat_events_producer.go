package producer

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	pkafka "github.com/beatlabs/patron/client/kafka/v2"
	"github.com/beatlabs/patron/log"
	"github.com/taxibeat/beatevents/gosdk"
	"github.com/taxibeat/beatevents/gosdk/entity"
	"github.com/taxibeat/beatevents/gosdk/protocol/kafka"
	"google.golang.org/protobuf/proto"
)

// BeatEventProducer is the pkafka producer.
type BeatEventProducer struct {
	schemaVersion string
	serviceName   string
	stack         string
	chError       <-chan error
	producer      *pkafka.AsyncProducer
}

// NewBeatEventProducer creates a new Kafka producer and connects to Kafka.
func NewBeatEventProducer(kafkaBrokers []string, stack string, schemaVersion string, serviceName string, config *sarama.Config) (*BeatEventProducer, error) {
	if config == nil {
		return nil, errors.New("error no sarama config provided")
	}

	pr, chErr, err := pkafka.New(kafkaBrokers, config).CreateAsync()
	if err != nil {
		return nil, err
	}

	go logErrors(chErr)
	return &BeatEventProducer{
		schemaVersion: schemaVersion,
		serviceName:   serviceName,
		stack:         stack,
		producer:      pr,
		chError:       chErr,
	}, nil
}

func logErrors(chErr <-chan error) {
	for {
		err := <-chErr
		topic := "unknown_topic"
		err = errors.Unwrap(err)
		if err != nil {
			var producerError *sarama.ProducerError
			ok := errors.Is(err, producerError)
			if ok {
				topic = producerError.Msg.Topic
			}
		}
		log.Errorf("error producing Kafka message for topic '%v': %v", topic, err)
	}
}

var (
	errInvalidTopic      = errors.New("invalid topic")
	errInvalidKey        = errors.New("invalid key")
	errInvalidEntityName = errors.New("invalid entityName")
	errInvalidMsg        = errors.New("invalid message")
)

// SendMessageWithKey sends a message to a certain topic with the provided key
func (p *BeatEventProducer) SendMessageWithKey(ctx context.Context, topic string, entityName string, key string, msg proto.Message) error {
	if topic == "" {
		return errInvalidTopic
	}

	if key == "" {
		return errInvalidKey
	}

	if entityName == "" {
		return errInvalidEntityName
	}
	if msg == nil {
		return errInvalidMsg
	}

	event, err := entity.New(p.stack, p.serviceName, entityName, gosdk.Proto(msg), gosdk.DataSchema(p.schemaVersion))
	if err != nil {
		return fmt.Errorf("could not create gosdk beat event: %w", err)
	}
	kafkaMsg := kafka.Create(topic, event)
	err = p.producer.Send(ctx, kafkaMsg)
	return err
}

// SendMessage sends a message to a certain topic
func (p *BeatEventProducer) SendMessage(ctx context.Context, topic string, entityName string, msg proto.Message) error {
	if topic == "" {
		return errInvalidTopic
	}

	if entityName == "" {
		return errInvalidEntityName
	}

	if msg == nil {
		return errInvalidMsg
	}

	event, err := entity.New(p.stack, p.serviceName, entityName, gosdk.Proto(msg), gosdk.DataSchema(p.schemaVersion))
	if err != nil {
		return fmt.Errorf("could not create gosdk beat event: %w", err)
	}
	kafkaMsg := kafka.Create(topic, event)
	err = p.producer.Send(ctx, kafkaMsg)
	return err
}

// Close terminates the kafka connection.
func (p *BeatEventProducer) Close() error {
	return p.producer.Close()
}
