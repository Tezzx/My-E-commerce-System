package database

import (
	"fmt"
	"net/url"
	"order-payment-system/config"

	"github.com/streadway/amqp"
)

// 创建 RabbitMQ 连接
func InitRabbitMQ(cfg *config.RabbitMQConfig) (*amqp.Connection, error) {
	password := url.QueryEscape(cfg.Password)
	vhost := url.QueryEscape(cfg.VHost)

	mqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.User,
		password,
		cfg.Host,
		cfg.Port,
		vhost,
	)

	conn, err := amqp.Dial(mqURL)
	if err != nil {
		return nil, err
	}
	return conn, err

}

// 声明一个队列（如果不存在则创建）
func DeclareQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclare(
		queueName, // 队列名称
		true,      // durable（持久化）
		false,     // autoDelete（自动删除）
		false,     // exclusive（排他性）
		false,     // noWait
		nil,       // arguments
	)
	return err
}
