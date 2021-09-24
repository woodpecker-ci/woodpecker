package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPubsub(t *testing.T) {
	var (
		wg sync.WaitGroup

		testTopic   = "test"
		testMessage = Message{
			Data: []byte("test"),
		}
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	broker := New()
	broker.Create(ctx, testTopic)
	go func() {
		broker.Subscribe(ctx, testTopic, func(message Message) { wg.Done() })
	}()
	go func() {
		broker.Subscribe(ctx, testTopic, func(message Message) { wg.Done() })
	}()

	<-time.After(500 * time.Millisecond)

	if _, ok := broker.(*publisher).topics[testTopic]; !ok {
		t.Errorf("Expect topic registered with publisher")
	}

	wg.Add(2)
	go func() {
		broker.Publish(ctx, testTopic, testMessage)
	}()

	wg.Wait()
	cancel()
}

func TestPublishNotFound(t *testing.T) {
	var (
		testTopic   = "test"
		testMessage = Message{
			Data: []byte("test"),
		}
	)
	broker := New()
	err := broker.Publish(context.Background(), testTopic, testMessage)
	if err != ErrNotFound {
		t.Errorf("Expect Not Found error when topic does not exist")
	}
}

func TestSubscribeNotFound(t *testing.T) {
	var (
		testTopic    = "test"
		testCallback = func(message Message) {}
	)
	broker := New()
	err := broker.Subscribe(context.Background(), testTopic, testCallback)
	if err != ErrNotFound {
		t.Errorf("Expect Not Found error when topic does not exist")
	}
}

func TestSubscriptionClosed(t *testing.T) {
	var (
		wg sync.WaitGroup

		testTopic    = "test"
		testCallback = func(Message) {}
	)

	broker := New()
	broker.Create(context.Background(), testTopic)
	go func() {
		broker.Subscribe(context.Background(), testTopic, testCallback)
		wg.Done()
	}()

	<-time.After(500 * time.Millisecond)

	if _, ok := broker.(*publisher).topics[testTopic]; !ok {
		t.Errorf("Expect topic registered with publisher")
	}

	wg.Add(1)
	broker.Remove(context.Background(), testTopic)
	wg.Wait()

	if _, ok := broker.(*publisher).topics[testTopic]; ok {
		t.Errorf("Expect topic removed from publisher")
	}
}
