package gcp

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// This function initializes Google PubSub client.
func (g *Gcp) pubsubClient() *pubsub.Client {
	ctx := context.Background()
	jwt, err := getJwtConfig(g.keyByte, g.scope)
	if err != nil {
		log.Fatalf("could not get JWT config with scope %s <%v>.", g.scope, err)
	}
	client, err := pubsub.NewClient(ctx, g.projectId, option.WithTokenSource(jwt.TokenSource(ctx)))
	if err != nil {
		log.Fatalf("could not initialize PubSub client <%v>", err)
	}

	return client
}
