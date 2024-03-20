package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
)

// This function initializes Google PubSub client.
func (g *Gcp) pubsubClient() {
	if g.pubsub == nil {
		ctx := context.Background()

		var err error
		var client *pubsub.Client
		var options []option.ClientOption
		var jwt *jwt.Config

		if g.emulatorHost != "" {
			os.Setenv("PUBSUB_EMULATOR_HOST", g.emulatorHost)
			// Emulators has no capability to authenticate
			options = append(options, option.WithoutAuthentication())
		} else {
			jwt, err = getJwtConfig(g.keyByte, g.scope)
			if err != nil {
				log.Fatalf("could not get JWT config with scope %s <%v>.", g.scope, err)
			}
			options = append(options, option.WithTokenSource(jwt.TokenSource(ctx)))
		}

		client, err = pubsub.NewClient(ctx, g.projectId, options...)
		if err != nil {
			log.Fatalf("could not initialize PubSub client <%v>", err)
		}

		g.pubsub = client
	}
}

func (g *Gcp) PubsubTopic(topic string) *pubsub.Topic {
	g.pubsubClient()
	return g.pubsub.Topic(topic)
}

func (g *Gcp) PubsubPublish(t *pubsub.Topic, message map[string]interface{}) (string, error) {
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

func (g *Gcp) PubsubSubscription(subscription string) *pubsub.Subscription {
	g.pubsubClient()
	return g.pubsub.Subscription(subscription)
}

func (g *Gcp) PubsubReceive(s *pubsub.Subscription, limit int, timeout int) ([]map[string]interface{}, error) {
	ctx := context.Background()
	s.ReceiveSettings.MaxOutstandingMessages = limit
	s.ReceiveSettings.MaxExtension = time.Duration(timeout) * time.Second
	var list []map[string]interface{}
	err := s.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		var message map[string]interface{}
		err := json.Unmarshal(m.Data, &message)
		if err != nil {
			log.Fatalf("unable to unmarshal subscription data <%v>", err)
		}
		list = append(list, message)
		m.Ack()
	})

	if err != nil {
		return nil, fmt.Errorf("unable to receive data from subscription %s <%v>", s, err)
	}

	return list, nil
}
