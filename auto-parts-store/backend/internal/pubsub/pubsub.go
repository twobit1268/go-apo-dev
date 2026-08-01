// Package pubsub wraps cloud.google.com/go/pubsub for the one event this
// app needs: OrderPlaced. Point PUBSUB_EMULATOR_HOST at the local emulator
// for dev/test; in Cloud Run, leave it unset and ADC talks to real Pub/Sub.
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
)

const (
	OrderEventsTopic          = "order-events"
	InventorySubscription     = "inventory-sub"
	NotificationsSubscription = "notifications-sub"
)

// OrderPlaced is published once per successful checkout.
type OrderPlaced struct {
	OrderID    string      `json:"orderId"`
	CustomerID string      `json:"customerId"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	PartID   string `json:"partId"`
	Quantity int    `json:"quantity"`
}

// Client owns a pubsub.Client and knows how to publish OrderPlaced events
// and ensure the topic/subscriptions it needs exist.
type Client struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	c, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub: new client: %w", err)
	}
	return &Client{client: c}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// EnsureTopology creates the order-events topic and its two subscriptions
// if they don't already exist. Safe to call repeatedly (e.g. on every
// process start) - it's idempotent.
func (c *Client) EnsureTopology(ctx context.Context) error {
	topic, err := ensureTopic(ctx, c.client, OrderEventsTopic)
	if err != nil {
		return err
	}
	c.topic = topic

	for _, subID := range []string{InventorySubscription, NotificationsSubscription} {
		if err := ensureSubscription(ctx, c.client, subID, topic); err != nil {
			return err
		}
	}
	return nil
}

func ensureTopic(ctx context.Context, client *pubsub.Client, id string) (*pubsub.Topic, error) {
	topic := client.Topic(id)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("pubsub: check topic %s: %w", id, err)
	}
	if ok {
		return topic, nil
	}
	topic, err = client.CreateTopic(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("pubsub: create topic %s: %w", id, err)
	}
	return topic, nil
}

func ensureSubscription(ctx context.Context, client *pubsub.Client, id string, topic *pubsub.Topic) error {
	sub := client.Subscription(id)
	ok, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: check subscription %s: %w", id, err)
	}
	if ok {
		return nil
	}
	_, err = client.CreateSubscription(ctx, id, pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		return fmt.Errorf("pubsub: create subscription %s: %w", id, err)
	}
	return nil
}

// Publisher is the narrow interface the checkout service depends on, so
// unit tests can supply a fake instead of a real Pub/Sub client.
type Publisher interface {
	PublishOrderPlaced(ctx context.Context, event OrderPlaced) error
}

func (c *Client) PublishOrderPlaced(ctx context.Context, event OrderPlaced) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("pubsub: marshal event: %w", err)
	}
	result := c.topic.Publish(ctx, &pubsub.Message{Data: data})
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: publish: %w", err)
	}
	return nil
}

// Subscribe blocks, invoking handler for every message on subID until ctx
// is cancelled. Used by cmd/worker for both the inventory and notifications
// subscribers.
func (c *Client) Subscribe(ctx context.Context, subID string, handler func(context.Context, OrderPlaced) error) error {
	sub := c.client.Subscription(subID)
	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var event OrderPlaced
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			// Malformed message: ack it anyway so it doesn't block the
			// subscription forever, but drop it rather than retry-looping.
			msg.Ack()
			return
		}
		if err := handler(ctx, event); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	})
}
