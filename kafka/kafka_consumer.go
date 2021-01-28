package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
	"time"
)

type KafkaHandler func(string, string, time.Time)

type Subscriber struct {
	cli      sarama.ConsumerGroup
	consumer consumer
	cancel   context.CancelFunc
}

func (s *Subscriber) Close() {
	s.cancel()
}

// consumer represents a Sarama consumer group consumer
type consumer struct {
	handler KafkaHandler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if nil != c.handler {
			c.handler(message.Topic, string(message.Value), message.Timestamp)
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func SubscribeKafka(topic []string, group string, addrs []string, handler KafkaHandler) (*Subscriber, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.ChannelBufferSize = 1024
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	cli, err := sarama.NewConsumerGroup(addrs, group, config)
	if err != nil {
		log.Error().Err(err).
			Strs("topic", topic).
			Str("group", group).
			Strs("addrs", addrs).
			Msgf("kafka sarama.NewConsumerGroup error")
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	consumer := consumer{
		handler: handler,
	}
	ret := &Subscriber{
		cli:      cli,
		consumer: consumer,
		cancel:   cancel,
	}
	go func() {
		for {
			if err := cli.Consume(ctx, topic, &consumer); err != nil {
				log.Error().Err(err).
					Strs("topic", topic).
					Str("group", group).
					Strs("addrs", addrs).
					Msgf("kafka consume error")
			}
			if ctx.Err() != nil {
				log.Error().Err(ctx.Err()).
					Strs("topic", topic).
					Str("group", group).
					Strs("addrs", addrs).
					Msgf("kafka subscriber cancel")
				return
			}
		}
	}()
	log.Info().Strs("topic", topic).
		Str("group", group).
		Strs("addrs", addrs).
		Msgf("create kafka subscriber")
	return ret, nil
}
