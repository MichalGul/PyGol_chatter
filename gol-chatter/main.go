package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const SimpleQueueDurable = "durable"
const SimpleQueueTransient = "transient"

// routing
const ExchangePyGol = "llm.dialog.exchange"
const RoutingKeyPyToGol = "py.to.gol"
const RoutingKeyGolToPy = "gol.to.py"

const QueuePyToGol = "q.py.to.gol"
const QueueGolToPy = "q.gol.to.py"


// Define a struct to match the JSON message structure
type Message struct {
    ConversationID string `json:"conversation_id"`
    Turn           int    `json:"turn"`
    MaxTurns       int    `json:"max_turns"`
    Sender         string `json:"sender"`
    Message        string `json:"message"`
}

func handleMessage(body []byte) (Message, error) {
	// Here you would unmarshal the JSON and process the message
	var msg Message
	err := json.Unmarshal(body, &msg)
	if err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return Message{}, err
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
			QueuePyToGol,
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

func main() {

	// TODO read from env file
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	channel, queue, err := declareAndBindQueue(conn, ExchangePyGol, QueuePyToGol, RoutingKeyPyToGol, SimpleQueueDurable)
	if err != nil {
		log.Fatal(err)
	}

	// declare response queue to send back to python
	_, _, err = declareAndBindQueue(conn, ExchangePyGol, QueueGolToPy, RoutingKeyGolToPy, SimpleQueueDurable)


	deliveryChannel, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	go func() error {
		defer channel.Close()

		for msg := range deliveryChannel {

			structMessage, _ := handleMessage(msg.Body)
			// TODO : process message here unmarshalling, calling LLM, etc.

			// todo send reply back to python bo to queue q.gol.to.py
			
			structMessage.Message = "Go is responding nicely, increasing turn number"
			structMessage.Sender = "golang"
			structMessage.Turn += 1

			responseBody, err := json.Marshal(structMessage)
			
			if err != nil {
				fmt.Printf("error marshalling response message: %v\n", err)
				return err
			}

			response_msg := amqp.Publishing{
				ContentType: "application/json",
				Body: responseBody,
				DeliveryMode: amqp.Persistent,
			} 
			
			publishErr := channel.PublishWithContext(context.Background(), ExchangePyGol, RoutingKeyGolToPy,false, false, response_msg)
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
