package kafka

import "github.com/IBM/sarama"

type Producer struct {
	syncProducer sarama.SyncProducer
}

func Init(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{syncProducer: producer}, nil
}

func (p *Producer) SendMessage(topic string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	_, _, err := p.syncProducer.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
