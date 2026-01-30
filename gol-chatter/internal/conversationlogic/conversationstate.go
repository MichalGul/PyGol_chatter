package conversationlogic



// Define a struct to match the JSON message structure
type Message struct {
    ConversationID string `json:"conversation_id"`
    Turn           int    `json:"turn"`
    MaxTurns       int    `json:"max_turns"`
    Sender         string `json:"sender"`
    Message        string `json:"message"`
}

type ConversationState struct {
	ConversationID string
	Actor         string
	CurrentTurn    int
	MaxTurns       int
}


func NewConversationState(conversationID, actorName string, maxTurns int) *ConversationState {
	return &ConversationState{
		ConversationID: conversationID,
		Actor:          actorName,
		CurrentTurn:    0,
		MaxTurns:       maxTurns,
	}
}

func (cs *ConversationState) IncrementTurn() {
	cs.CurrentTurn++
}

func (cs *ConversationState) UpdateTurn(turn int) {
	cs.CurrentTurn = turn
}

func (cs *ConversationState) IsConversationOver() bool {
	return cs.CurrentTurn >= cs.MaxTurns
}
