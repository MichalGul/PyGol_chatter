package routing

// Define a struct to match the JSON message structure
type Message struct {
    ConversationID string `json:"conversation_id"`
    Turn           int    `json:"turn"`
    MaxTurns       int    `json:"max_turns"`
    Sender         string `json:"sender"`
    Message        string `json:"message"`
}