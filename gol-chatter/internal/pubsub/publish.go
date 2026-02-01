package pubsub
// Logic for publishing messages to RabbitMQ

import (
	"encoding/json"
	"fmt"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishMesssageJSON publishes a message of generic type T as JSON to the specified exchange and routing key
func PublishMesssageJSON[T any](channel *amqp.Channel, exchangeName, routingKey string, message T) error {


	messageBody, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("error marshalling message: %v\n", err)
		return err
	}

	amqpMessage := amqp.Publishing{
		ContentType: "application.json",
		Body: 	  messageBody,
		DeliveryMode: amqp.Persistent,
	}

	// Publish the response message back to python to exchange with routing key gol.to.py
	publishErr := channel.PublishWithContext(context.Background(), exchangeName, routingKey, false, false, amqpMessage)
	if publishErr != nil {
		fmt.Printf("error publishing message to queue: %v\n", publishErr)
		return fmt.Errorf("nac: %w", publishErr)
	}

	return nil

}