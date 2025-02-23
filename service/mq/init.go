package mq

import (
	"easy-chat/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitMQChannel *amqp.Channel

func Init() error {
	url := buildURL()
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	rabbitMQChannel, err = conn.Channel()
	if err != nil {
		return err
	}

	_, err = rabbitMQChannel.QueueDeclare(
		chatRequestQueue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for range chatRequestConsumerNum {
		go startChatRequestConsumer()
	}

	return nil
}

func buildURL() string {
	cfg := config.Get()
	username := cfg.MQ.Username
	password := cfg.MQ.Password
	port := cfg.MQ.Port
	return "amqp://" + username + ":" + password + "@localhost:" + port + "/"
}
