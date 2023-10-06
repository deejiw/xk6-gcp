package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// This function initializes Google PubSub client.
func (g *Gcp) pubsubClient() {
	if g.pubsub == nil {
		ctx := context.Background()
		jwt, err := getJwtConfig(g.keyByte, g.scope)
		if err != nil {
			log.Fatalf("could not get JWT config with scope %s <%v>.", g.scope, err)
		}
		client, err := pubsub.NewClient(ctx, g.projectId, option.WithTokenSource(jwt.TokenSource(ctx)))
		if err != nil {
			log.Fatalf("could not initialize PubSub client <%v>", err)
		}

		g.pubsub = client
	}
}

func (g *Gcp) GetTopic(topic string) *pubsub.Topic {
	g.pubsubClient()
	return g.pubsub.Topic(topic)
}

func (g *Gcp) Publish(t *pubsub.Topic, message map[string]interface{}) (string, error) {
	ctx := context.Background()

	b, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON <%v>", err)
	}

	res := t.Publish(ctx, &pubsub.Message{Data: b})

	msgId, err := res.Get(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to get message ID <%v>", err)
	}

	return msgId, nil
}

func (g *Gcp) GetSubscription(subscription string) *pubsub.Subscription {
	g.pubsubClient()
	return g.pubsub.Subscription(subscription)
}

func (g *Gcp) Receive(s *pubsub.Subscription) ([]map[string]interface{}, error) {
	ctx := context.Background()
	list := []map[string]interface{}{}
	err := s.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		var data map[string]interface{}
		err := json.Unmarshal(m.Data, &data)

		if err != nil {
			log.Fatalf("unable to unmarshal subscription data <%v>", err)
		}

		list = append(list, data)
		m.Ack()
	})

	if err != nil {
		return nil, fmt.Errorf("unable to receive data from subscription %s <%v>", s, err)
	}

	return list, nil
}
