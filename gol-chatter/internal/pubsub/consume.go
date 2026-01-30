package pubsub

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const SimpleQueueDurable = "durable"
const SimpleQueueTransient = "transient"

// DeclareAndBindQueue declares a queue and binds it to an exchange with a routing key
func DeclareAndBindQueue(
	conn *amqp.Connection,
	exchangeName,
	queueName,
	routingKey string,
	queueType simpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {

	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, errors.New("Error creating Channel")
	}

	declaredQueue, err := channel.QueueDeclare(
		queueName,
		queueType == "durable",
		queueType == "transient",
		queueType == "transient",
		false,
		nil) // change later to deade letter exchange

	if err != nil {
		log.Println(err)
		return channel, amqp.Queue{}, errors.New("Error declaring queue")
	}

	// Bind the queue to the exchange with the routing key
	errBind := channel.QueueBind(queueName, routingKey, exchangeName, false, nil)
	if errBind != nil {
		errorMsg := fmt.Sprintf("Error binding queue %s to exchange %s", queueName, exchangeName)
		return channel, amqp.Queue{}, errors.New(errorMsg)
	}

	return channel, declaredQueue, nil

}

// Abstracted function to subscribe to a queue and process JSON messages
// We will pass handler function that processes the message and accepts generic type T like conversationlogic.ConversationState
// It will Declare and Binde the queue and start consuming messages
// Each message will be json unmarshaled
func SubscribeToQueueJSON[T any](
	connection *amqp.Connection,
	exchangeName,
	queueName,
	routingKey string,
	queueType simpleQueueType,
	handler func(T) error,
) error {

	channel, queue, err := DeclareAndBindQueue(connection, exchangeName, queueName, routingKey, queueType)
	if err != nil {
		return errors.New("Error declaring and binding queue " + queueName + "to exchange " + exchangeName)
	}

	// Consume messages form the declared queue
	deliveryChannel, err := channel.Consume(queue.Name, "", false, false, false, false, nil)

	// Run a goroutine to process messages
	go func() error {
		defer channel.Close()

		for msg := range deliveryChannel {

			var messageBody T

			marshallErr := json.Unmarshal(msg.Body, &messageBody)
			if marshallErr != nil {
				return fmt.Errorf("error unmarshalling message %v", marshallErr)
			}

			processErr := handler(messageBody)
			if processErr != nil {
				return fmt.Errorf("error processing message %v", processErr)
			}

			// TODO: handle Ack, NackRequeue and Nackdiscard based on processing result
			// For now Acknowledge the message after processing no mather if error or not
			msg.Ack(false)


		}
		return nil

	}()

	return nil
}
