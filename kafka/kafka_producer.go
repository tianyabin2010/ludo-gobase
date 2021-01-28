package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
)

type Producer struct {
	producer sarama.AsyncProducer
}

func (p *Producer) Close() error {
	if nil != p && nil != p.producer {
		return p.producer.Close()
	}
	return nil
}

func (p *Producer) Post(topic, msg string) {
	if nil != p && nil != p.producer {
		p.producer.Input() <- &sarama.ProducerMessage{
			Topic:     topic,
			Value:     sarama.StringEncoder(msg),
		}
	}
}

func NewProducer(addrs []string) (*Producer, error) {
	conf := sarama.NewConfig()
	conf.ChannelBufferSize = 1024
	conf.Version = sarama.V0_11_0_0
	producer, err := sarama.NewAsyncProducer(addrs, conf)
	if err != nil {
		log.Error().Err(err).
			Strs("addrs", addrs).
			Msgf("kafka sarama.NewAsyncProducer error")
		return nil, err
	}
	ret := &Producer{
		producer: producer,
	}
	return ret, nil
}