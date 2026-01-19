package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"errors"

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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	channel, queue, err := declareAndBindQueue(conn, ExchangePyGol, QueuePyToGol, RoutingKeyPyToGol, SimpleQueueDurable)
	if err != nil {
		log.Fatal(err)
	}

	deliveryChannel, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	// forever := make(chan bool)
	go func() error {
		defer channel.Close()

		for msg := range deliveryChannel {
			log.Printf(" [x] Received %s", msg.Body)
			// TODO : process message here unmarshalling, calling LLM, etc.

			// todo send reply back to python bo to queue q.gol.to.py

			// reply := []byte("Pong from Go to: " + string(d.Body))
			// err = d.Channel.Publish("", q.Name, false, false, amqp.Publishing{
			//     DeliveryMode: amqp.Persistent,
			//     Body:         reply,
			// })
			// if err != nil {
			//     log.Printf("Failed to publish: %v", err)
			// }
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
