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

// 声明一个带死信队列的队列
func DeclareQueueWithDLX(ch *amqp.Channel, queueName string) error {
	// 1. 声明死信交换机
	dlxExchange := "dlx_" + queueName
	if err := ch.ExchangeDeclare(dlxExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// 2. 声明死信队列
	dlqName := queueName + "_dlq"
	if _, err := ch.QueueDeclare(dlqName, true, false, false, false, nil); err != nil {
		return err
	}

	// 3. 绑定死信队列
	if err := ch.QueueBind(dlqName, queueName+"_routing_key", dlxExchange, false, nil); err != nil {
		return err
	}

	// 4. 声明主队列（带死信参数）
	args := amqp.Table{
		"x-dead-letter-exchange":    dlxExchange,
		"x-dead-letter-routing-key": queueName + "_routing_key",
	}

	_, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args,
	)
	return err
}

// 声明延迟队列和对应的超时消费队列
func DeclareDelayTimeoutQueue(ch *amqp.Channel) error {
	// 1. 声明超时执行的交换机和队列
	timeoutExchange := "order_timeout_exchange"
	if err := ch.ExchangeDeclare(timeoutExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	timeoutQueue := "order_timeout_queue"
	if _, err := ch.QueueDeclare(timeoutQueue, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(timeoutQueue, "timeout", timeoutExchange, false, nil); err != nil {
		return err
	}

	// 2. 声明中间延迟队列（没有消费者，通过 TTL 让消息死信滑入超时队列）
	delayArgs := amqp.Table{
		"x-dead-letter-exchange":    timeoutExchange,
		"x-dead-letter-routing-key": "timeout",
		"x-message-ttl":             int32(600000),
	}
	_, err := ch.QueueDeclare("order_delay_queue", true, false, false, false, delayArgs)
	return err
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
	fmt.Printf("========================")
	return err
}
