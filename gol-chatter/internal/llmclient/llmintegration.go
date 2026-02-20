package llmclient 


type LLMClient interface {
	GetLLMSimpleResponse(prompt string) (string, error) // More to add
}

