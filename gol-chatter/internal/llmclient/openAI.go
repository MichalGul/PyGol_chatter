package llmclient

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"context"
	"errors"
)


type OpenAIClient struct {
    client *openai.Client
}


func CreateNewOpenAIClient(apiKey string) *OpenAIClient {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	
	return &OpenAIClient{client: &client}
}



func (oaiClient * OpenAIClient) GetLLMSimpleResponse(prompt string) (string, error) {

	// Use the client to send a request to the LLM API and get a response.
	// This will involve formatting the prompt correctly and handling the API response.
	resp, err := oaiClient.client.Responses.New(context.Background(), 
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