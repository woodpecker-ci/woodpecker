package pubsub

import (
	"testing"
	"time"
)

func TestTopicSubscribe(t *testing.T) {
	sub := new(subscriber)
	top := newTopic("foo")
	top.subscribe(sub)
	if _, ok := top.subs[sub]; !ok {
		t.Errorf("Expect subscription registered with topic on subscribe")
	}
}

func TestTopicUnsubscribe(t *testing.T) {
	sub := new(subscriber)
	top := newTopic("foo")
	top.subscribe(sub)
	if _, ok := top.subs[sub]; !ok {
		t.Errorf("Expect subscription registered with topic on subscribe")
	}
	top.unsubscribe(sub)
	if _, ok := top.subs[sub]; ok {
		t.Errorf("Expect subscription de-registered with topic on unsubscribe")
	}
}

func TestTopicClose(t *testing.T) {
	sub := new(subscriber)
	top := newTopic("foo")
	top.subscribe(sub)
	go func() {
		top.close()
	}()
	select {
	case <-top.done:
	case <-time.After(1 * time.Second):
		t.Errorf("Expect subscription closed")
	}
}
