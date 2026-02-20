package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/conversationlogic"
	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/routing"
	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/pubsub"
	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/llmclient"
	"github.com/openai/openai-go/v3"

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

		// Send response back to python. Marshalling happens in the publisher function
		responseMessage := conversationlogic.Message{
			ConversationID: msg.ConversationID,
			Turn:           convState.CurrentTurn,
			MaxTurns:       msg.MaxTurns,
			Sender:         convState.Actor,
			Message:        "Go is responding nicely, increasing turn number",
		}

		pubsub.PublishMesssageJSON(
			channel,
			routing.ExchangePyGol,
			routing.RoutingKeyGolToPy,
			responseMessage,
		)


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

		responseMessage := conversationlogic.Message{
			ConversationID: msg.ConversationID,
			Turn:           convState.CurrentTurn,
			MaxTurns:       msg.MaxTurns,
			Sender:         convState.Actor,
			Message:        userInput,
		}

		pubsub.PublishMesssageJSON(
			channel,
			routing.ExchangePyGol,
			routing.RoutingKeyGolToPy,
			responseMessage,
		)

		log.Printf("ACK message for conversation %s (turn %d/%d)", msg.ConversationID, convState.CurrentTurn, msg.MaxTurns)


		return nil

	}
}


func handleMessageLLMResponse(convState *conversationlogic.ConversationState, channel *amqp.Channel, client *openai.Client) func(conversationlogic.Message) error {
	return func(msg conversationlogic.Message) error {
				convState.UpdateTurn(msg.Turn)
		convState.IncrementTurn()
		log.Printf("Processing message from %s: %s (Turn %d/%d)", msg.Sender, msg.Message, msg.Turn, msg.MaxTurns)

		fmt.Printf("Message from py_chatter: %s\n", msg.Message)

		// Get response from LLM, todo give message turn to llm
		llmResponse, err := llmclient.GetLLMSimpleResponse(client, msg.Message)
		if err != nil {
			log.Printf("Error getting response from LLM: %v", err)
			llmResponse = "Error getting response from LLM"
		}

		responseMessage := conversationlogic.Message{
			ConversationID: msg.ConversationID,
			Turn:           convState.CurrentTurn,
			MaxTurns:       msg.MaxTurns,
			Sender:         convState.Actor,
			Message:        llmResponse,
		}

		pubsub.PublishMesssageJSON(
			channel,
			routing.ExchangePyGol,
			routing.RoutingKeyGolToPy,
			responseMessage,
		)

		log.Printf("ACK message for conversation %s (turn %d/%d)", msg.ConversationID, convState.CurrentTurn, msg.MaxTurns)

		return nil
	}
}