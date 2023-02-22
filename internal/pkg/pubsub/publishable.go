package pubsub

type Publishable interface {
	GetEventTopicName() string
}
