package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
	"time"
)

func InitKafka(addrs ...string) error {
	config := sarama.NewConfig()
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	consumer, err := sarama.NewConsumer(addrs, config)
	if err != nil {
		log.Error().Err(err).Strs("addrs", addrs).
			Msgf("InitKafka NewConsumer error")
		return err
	}
	client, err := sarama.NewClient(addrs, config)
	if err != nil {
		panic("client create error")
	}
	defer client.Close()
}