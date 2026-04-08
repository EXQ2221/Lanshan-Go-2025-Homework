// internal/dal/kafka/consumer.go
package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

// Consumer 消费组消费者
type Consumer struct{}

func NewConsumer() *Consumer {
	return &Consumer{}
}

// Setup 在 session 开始时调用
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup 在 session 结束时调用
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim 每拿到一个 partition 就调一次
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// 处理消息
		var event SecurityEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("unmarshal error: %v", err)
			session.MarkMessage(msg, "") // 标记已消费，避免卡住
			continue
		}

		// 根据事件类型处理
		switch event.EventType {
		case EventTypeAbnormalGeoLogin:
			log.Printf("[异地登录] user=%d %s -> %s ip=%s",
				event.UserID, event.PreviousCity, event.CurrentCity, event.IP)
			// TODO: 存数据库 / 发通知等
		default:
			log.Printf("[安全事件] type=%s user=%d", event.EventType, event.UserID)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

// StartConsumer 启动消费（阻塞）
func StartConsumer(brokers []string, groupID string, topics []string) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0                      // Kafka 版本
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 从最早开始读

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	consumer := NewConsumer()

	for {
		if err := client.Consume(context.Background(), topics, consumer); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}
}
