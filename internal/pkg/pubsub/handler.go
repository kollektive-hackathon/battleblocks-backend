package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
)

type SubscriptionHandler struct {
	SubscriptionId string
	Handler        func(ctx context.Context, message *pubsub.Message)
}
