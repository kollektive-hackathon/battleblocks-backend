package blockchain

import "github.com/google/uuid"

type Authorizer struct {
	KmsResourceId        string `json:"kmsResourceId"`
	ResourceOwnerAddress string `json:"resourceOwnerAddress"`
}

type Command struct {
	Id          string       `json:"id"`
	Type        string       `json:"type"`
	Payload     []any        `json:"payload"`
	Authorizers []Authorizer `json:"authorizers"`
}

func (bc Command) GetEventTopicName() string {
	return "blockchain.flow.commands"
}

func NewBlockchainCommand(commandType string, payload []any, authorizers []Authorizer) Command {
	return Command{
		Id:          uuid.New().String(),
		Type:        commandType,
		Payload:     payload,
		Authorizers: authorizers,
	}
}
