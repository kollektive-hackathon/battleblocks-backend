package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var ctx context.Context
var client *pubsub.Client

func InitPubSub() {
	projectID := viper.Get("GOOGLE_PROJECT_ID").(string)
	if projectID == "" {
		log.Fatal().Msg("Pub sub missing projectID to initialize")
	}
	fmt.Printf("Init pubsub with projectID:%v", projectID)
	ctx = context.Background()
	var err error
	client, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("Error initializing pub sub connection"))
	}
	log.Info().Msg(fmt.Sprintf("Successful pubsub init"))
}

func Subscribe(subscriptionHandler SubscriptionHandler) {
	sub := client.Subscription(subscriptionHandler.SubscriptionId)
	err := sub.Receive(ctx, subscriptionHandler.Handler)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("Subscriber error for sub id %s", subscriptionHandler.SubscriptionId))
	}
}

func Publish(message Publishable, options ...map[string]any) {
	t := getTopic(message.GetEventTopicName())
	defer t.Stop()

	result := t.Publish(ctx, &pubsub.Message{Data: encodeMessage(message)})

	go func(res *pubsub.PublishResult) {
		_, err := res.Get(ctx)
		if err != nil {
			log.Warn().Msg(fmt.Sprintf("Failed to publish message for %s", message.GetEventTopicName()))
			return
		}
	}(result)
}

func CloseClient() {
	client.Close()
}

func getTopic(topicName string) *pubsub.Topic {
	t := client.Topic(topicName)
	if t == nil {
		log.Info().Msg(fmt.Sprintf("Topic %s does not exist. Creating new", topicName))
		nt, err := client.CreateTopic(ctx, topicName)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Cant create topic %s", topicName))
		}
		return nt
	}
	return t
}

func encodeMessage(message any) []byte {
	switch message.(type) {
	case string:
		return []byte(message.(string))

	default:
		bytes, _ := json.Marshal(message)
		return bytes
	}
}
