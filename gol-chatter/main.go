package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/conversationlogic"
	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/pubsub"
	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/routing"
	"github.com/MichalGul/PyGol/PyGol_chatter/gol-chatter/internal/llmclient"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func loadEnvFile(path string) error {

	err := godotenv.Load(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("No env file found at %s, proceeding with system environment variables")
			return nil
		}
		return fmt.Errorf("load env file: %w", err)
	}

	return nil
}

func main() {

	if err := loadEnvFile(".env"); err != nil {
		log.Fatalf("failed to load env file: %v", err)
	}
	amqpURL := os.Getenv("RABBITMQ_URL")

	if amqpURL == "" {
		log.Fatalf("RabbitMQ URL not set in environment variable RABBITMQ_URL. Exit.")
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conversationState := conversationlogic.NewConversationState("conv1", "Gol-chatter", 10)
	fmt.Printf("Initialized conversation state: %+v\n", conversationState)


    // Use the interface to get a response
    var llm llmclient.LLMClient

	llmType := os.Getenv("LMM_CLIENT_TYPE")
	if llmType == "" {
		log.Printf("Warning: LMM_CLIENT_TYPE not set in environment variables. Defaulting to openai.")
		llmType = "openai"
	}

	if llmType == "openai" {
		log.Printf("Using OpenAI as LLM client.")

		llmAPIKey := os.Getenv("OPENAI_API_KEY")
		if llmAPIKey == "" {
			log.Printf("Warning: OPENAI_API_KEY not set in environment variables. LLM integration will not work.")
		}

		// TODO interface for client types
		llmClient := llmclient.CreateNewOpenAIClient(llmAPIKey)
		fmt.Printf("Initialized llm OpenAI client: \n")
		llm = llmClient

	} else { // todo add more LLM clients as needed
		log.Printf("Error: Unsupported LLM_CLIENT_TYPE '%s'. No LLM integration will be available.", llmType)
	}




	// Declare queue to consume from python todo not sure if this requred?
	channel, _, err := pubsub.DeclareAndBindQueue(
		conn,
		routing.ExchangePyGol,
		routing.QueuePyToGol,
		routing.RoutingKeyPyToGol,
		pubsub.SimpleQueueDurable)

	if err != nil {
		log.Fatal(err)
	}

	// declare and bind to exchange response queue to send back to python
	// creates queue q.gol.to.py that exchanges messages with routing key gol.to.py
	_, _, err = pubsub.DeclareAndBindQueue(
		conn,
		routing.ExchangePyGol,
		routing.QueueGolToPy,
		routing.RoutingKeyGolToPy,
		pubsub.SimpleQueueDurable)

	if err != nil {
		log.Fatal(err)
	}


	// Start subscribing to the queue with a handler function
	// It runs goroutine that processes messages as they arrive
	// err = pubsub.SubscribeToQueueJSON[conversationlogic.Message](
	// 	conn,
	// 	routing.ExchangePyGol,
	// 	routing.QueuePyToGol,
	// 	routing.RoutingKeyPyToGol,
	// 	pubsub.SimpleQueueDurable,
	// 	handleMessageManualResponse(conversationState, channel),
	// )

		err = pubsub.SubscribeToQueueJSON[conversationlogic.Message](
		conn,
		routing.ExchangePyGol,
		routing.QueuePyToGol,
		routing.RoutingKeyPyToGol,
		pubsub.SimpleQueueDurable,
		handleMessageLLMResponse(conversationState, channel, llm),
	)


	log.Printf(" [*] Waiting for messages...")
	// <-forever	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")

}
