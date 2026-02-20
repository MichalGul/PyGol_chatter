package llmclient 


import (
	"context"
	"errors"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// Start with openAi integration, then add more LLMs as needed.
// TODO make own type with interface to abstract away from specific LLM client implementations and allow for multiple LLMs to be used interchangeably. This will also make it easier to mock the client for testing.
func CreateClient(apiKey string) *openai.Client {
	// Create a client for the LLM API (e.g., OpenAI, Perplexity, etc.)
	// This will likely involve setting up authentication and any necessary configuration.

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	
	return &client
}

func GetLLMSimpleResponse(client *openai.Client, prompt string) (string, error) {

	// Use the client to send a request to the LLM API and get a response.
	// This will involve formatting the prompt correctly and handling the API response.
	resp, err := client.Responses.New(context.Background(), 
	responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		Model: openai.ChatModelGPT4oMini,
		Instructions: openai.String("You are a helpful assistant, that provides concise and informative responses to the user's input. Please respond to the user's message in a clear and helpful manner. Be short an consice."),
	})

	if err != nil {
		return "", errors.New("Error getting response from LLM API: " + err.Error())
	}

	// TODO handle response ID to create conversation . lookup resp fields because there are a lot of them
	return resp.OutputText(), nil

}

// Todo handle conversation history and context check client.Conversations.New