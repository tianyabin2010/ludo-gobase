package kafka

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
	"time"
)

type KafkaHandler func(string, string, time.Time)
type KafkaByteHandler func(string, []byte, time.Time)

type Subscriber struct {
	cli      sarama.ConsumerGroup
	consumer sarama.ConsumerGroupHandler
	cancel   context.CancelFunc
}

func (s *Subscriber) Close() {
	s.cancel()
}

type consumer struct {
	handler KafkaHandler
}

func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if nil != c.handler {
			c.handler(message.Topic, string(message.Value), message.Timestamp)
		}
		session.MarkMessage(message, "")
	}
	return nil
}

type byte_consumer struct {
	handler KafkaByteHandler
}

func (c *byte_consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *byte_consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *byte_consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if nil != c.handler {
			c.handler(message.Topic, message.Value, message.Timestamp)
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func SubscribeKafka(topic []string, group string, addrs []string, handler KafkaHandler) (*Subscriber, error) {
	if nil == addrs || len(addrs) <= 0 {
		return nil, errors.New("kafka subscribe error, addrs is nil")
	}
	if nil == topic || len(topic) <= 0 {
		return nil, errors.New("kafka subscribe error, topic is nil")
	}
	if nil == handler {
		log.Error().Strs("topics", topic).
			Strs("addrs", addrs).
			Msgf("SubscribeKafkaV2 handler is nil")
	}
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
	consumer := &consumer{
		handler: handler,
	}
	ret := &Subscriber{
		cli:      cli,
		consumer: consumer,
		cancel:   cancel,
	}
	go func() {
		for {
			if err := cli.Consume(ctx, topic, consumer); err != nil {
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

func SubscribeKafkaV2(topic []string, group string, addrs []string, handler KafkaByteHandler) (*Subscriber, error) {
	if nil == addrs || len(addrs) <= 0 {
		return nil, errors.New("kafka subscribe error, addrs is nil")
	}
	if nil == topic || len(topic) <= 0 {
		return nil, errors.New("kafka subscribe error, topic is nil")
	}
	if nil == handler {
		log.Error().Strs("topics", topic).
			Strs("addrs", addrs).
			Msgf("SubscribeKafkaV2 handler is nil")
	}
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
	consumer := &byte_consumer{
		handler: handler,
	}
	ret := &Subscriber{
		cli:      cli,
		consumer: consumer,
		cancel:   cancel,
	}
	go func() {
		for {
			if err := cli.Consume(ctx, topic, consumer); err != nil {
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