package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("llm.dialog.exchange", "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	q, err := ch.QueueDeclare("q.py.to.gol", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	deliveryChannel, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	// forever := make(chan bool)
	go func() {
		for d := range deliveryChannel {
			log.Printf(" [x] Received %s", d.Body)
			// reply := []byte("Pong from Go to: " + string(d.Body))
			// err = d.Channel.Publish("", q.Name, false, false, amqp.Publishing{
			//     DeliveryMode: amqp.Persistent,
			//     Body:         reply,
			// })
			// if err != nil {
			//     log.Printf("Failed to publish: %v", err)
			// }
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages...")
	// <-forever	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")

}
