package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/conversationlogic"
	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Handler that just logs the message
func handleMessageBasicResponse(convState *conversationlogic.ConversationState, channel *amqp.Channel) func(conversationlogic.Message) error {

	// Unmarshalling happens in the generic subscriber function
	return func(msg conversationlogic.Message) error {
		convState.UpdateTurn(msg.Turn)
		convState.IncrementTurn()

		log.Printf("Current turn incremented by %s: %d", convState.Actor, convState.CurrentTurn)
		log.Printf("Processing message from %s: %s (Turn %d/%d)", msg.Sender, msg.Message, msg.Turn, msg.MaxTurns)

		// Send response back to python
		// TODO wrap into a publisher function
		responseMessage := conversationlogic.Message{
			ConversationID: msg.ConversationID,
			Turn:           convState.CurrentTurn,
			MaxTurns:       msg.MaxTurns,
			Sender:         convState.Actor,
			Message:        "Go is responding nicely, increasing turn number",
		}

		responseBody, err := json.Marshal(responseMessage)
		if err != nil {
			fmt.Printf("error marshalling response message: %v\n", err)
			return err
		}

		responseMsg := amqp.Publishing{
			ContentType:  "application/json",
			Body:         responseBody,
			DeliveryMode: amqp.Persistent,
		}

		// Publish the response message back to python to exchange with routing key gol.to.py
		publishErr := channel.PublishWithContext(context.Background(), routing.ExchangePyGol, routing.RoutingKeyGolToPy, false, false, responseMsg)
		if publishErr != nil {
			fmt.Printf("error publishing message to queue: %v\n", publishErr)
			return fmt.Errorf("nac: %w", publishErr)
		}

		log.Printf("ACK message for conversation %s (turn %d/%d)", msg.ConversationID, convState.CurrentTurn, msg.MaxTurns)
		return nil
	}
}

func handleMessageManualResponse(convState *conversationlogic.ConversationState, channel *amqp.Channel) func(conversationlogic.Message) error {
	return func(msg conversationlogic.Message) error {
		convState.UpdateTurn(msg.Turn)
		convState.IncrementTurn()
		log.Printf("Processing message from %s: %s (Turn %d/%d)", msg.Sender, msg.Message, msg.Turn, msg.MaxTurns)

		fmt.Printf("Message from py_chatter: %s\n", msg.Message)

		inputScanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Response time > \n")
		inputScanner.Scan()
		err := inputScanner.Err()
		userInput := ""
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
		userInput = inputScanner.Text()
		if userInput == "" {
			userInput = "No response provided from Go chatter"
		}

		// TODO wrap into a publisher function
		responseMessage := conversationlogic.Message{
			ConversationID: msg.ConversationID,
			Turn:           convState.CurrentTurn,
			MaxTurns:       msg.MaxTurns,
			Sender:         convState.Actor,
			Message:        userInput,
		}

		responseBody, err := json.Marshal(responseMessage)
		if err != nil {
			fmt.Printf("error marshalling response message: %v\n", err)
			return err
		}

		responseMsg := amqp.Publishing{
			ContentType:  "application/json",
			Body:         responseBody,
			DeliveryMode: amqp.Persistent,
		}

		// Publish the response message back to python to exchange with routing key gol.to.py
		publishErr := channel.PublishWithContext(context.Background(), routing.ExchangePyGol, routing.RoutingKeyGolToPy, false, false, responseMsg)
		if publishErr != nil {
			fmt.Printf("error publishing message to queue: %v\n", publishErr)
			return fmt.Errorf("nac: %w", publishErr)
		}

		return nil

	}
}
