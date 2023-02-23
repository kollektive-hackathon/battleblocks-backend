package blockchain

import "github.com/google/uuid"

type Command struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Payload []any  `json:"payload"`
}

func (bc Command) GetEventTopicName() string {
	return "blockchain.flow.commands"
}

func NewBlockchainCommand(commandType string, payload []any) Command {
	return Command{
		Id:      uuid.New().String(),
		Type:    commandType,
		Payload: payload,
	}
}
