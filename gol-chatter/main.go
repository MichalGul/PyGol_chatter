package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const SimpleQueueDurable = "durable"
const SimpleQueueTransient = "transient"

// Nothing handler for now
func handleMessage(body []byte) (routing.Message, error) {
	// Here you would unmarshal the JSON and process the message
	var msg routing.Message
	err := json.Unmarshal(body, &msg)
	if err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return routing.Message{}, err
	}

	// Process the message (for now, just print it)
	log.Printf("Processing message from %s: %s (Turn %d/%d)", msg.Sender, msg.Message, msg.Turn, msg.MaxTurns)

	return msg, nil
}

// Todo move to internal module
func declareAndBindQueue(
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
		routing.QueuePyToGol,
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

func loadEnvFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("load env file: %w", err)
	}

	for _, raw := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if key == "" {
			continue
		}
		os.Setenv(key, value)
	}

	return nil
}

func main() {

	if err := loadEnvFile(".env"); err != nil {
		log.Fatalf("failed to load env file: %v", err)
	}

	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	channel, queue, err := declareAndBindQueue(conn, routing.ExchangePyGol, routing.QueuePyToGol, routing.RoutingKeyPyToGol, SimpleQueueDurable)
	if err != nil {
		log.Fatal(err)
	}

	// declare response queue to send back to python
	_, _, err = declareAndBindQueue(conn, routing.ExchangePyGol, routing.QueueGolToPy, routing.RoutingKeyGolToPy, SimpleQueueDurable)

	deliveryChannel, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	go func() error {
		defer channel.Close()

		for msg := range deliveryChannel {

			structMessage, _ := handleMessage(msg.Body)
			// TODO : process message here unmarshalling, calling LLM, etc.

			structMessage.Message = "Go is responding nicely, increasing turn number"
			structMessage.Sender = "golang"
			structMessage.Turn += 1

			responseBody, err := json.Marshal(structMessage)

			if err != nil {
				fmt.Printf("error marshalling response message: %v\n", err)
				return err
			}

			response_msg := amqp.Publishing{
				ContentType:  "application/json",
				Body:         responseBody,
				DeliveryMode: amqp.Persistent,
			}

			publishErr := channel.PublishWithContext(context.Background(), routing.ExchangePyGol, routing.RoutingKeyGolToPy, false, false, response_msg)
			if publishErr != nil {
				fmt.Printf("error publishing message to queue: %v\n", publishErr)
				return publishErr
			}

			msg.Ack(false)
		}

		return nil
	}()

	log.Printf(" [*] Waiting for messages...")
	// <-forever	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")

}
