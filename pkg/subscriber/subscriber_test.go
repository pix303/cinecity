package subscriber_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/cinecity/pkg/subscriber"
	"github.com/stretchr/testify/assert"
)

func TestNewSubscriptionState(t *testing.T) {
	slog.Info("start testing")
	state := subscriber.NewSubscriptionsState()
	assert.Equal(t, 0, state.NumSubscribers(), "initial num of subscribers must be null")
}

func TestAddSubscriber(t *testing.T) {
	state := subscriber.NewSubscriptionsState()
	subAddr := actor.NewAddress("local", "subscriber")
	state.AddSubscription(subAddr)
	assert.Equal(t, 1, state.NumSubscribers(), "num of subscribers after add must be 1")
}

func TestRemoveSubscriber(t *testing.T) {
	state := subscriber.NewSubscriptionsState()
	subAddr := actor.NewAddress("local", "subscriber")
	state.AddSubscription(subAddr)
	state.RemoveSubscription(subAddr)
	assert.Equal(t, 0, state.NumSubscribers(), "num of subscribers after add and remove must be 0")
}

func TestNumSubscriber(t *testing.T) {
	state := subscriber.NewSubscriptionsState()
	subAddr := actor.NewAddress("local", "subscriber")
	state.AddSubscription(subAddr)
	assert.Equal(t, 1, state.NumSubscribers(), "num of subscribers after add and remove must be 0")
}

func TestNotifySubscribers(t *testing.T) {
	state := subscriber.NewSubscriptionsState()
	
	subAddr1 := actor.NewAddress("local", "subscriber1")
	subAddr2 := actor.NewAddress("local", "subscriber2")
	
	state.AddSubscription(subAddr1)
	state.AddSubscription(subAddr2)
	
	fromAddr := actor.NewAddress("local", "sender")
	msg := actor.NewMessage(nil, fromAddr, "test message")
	
	state.NotifySubscribers(msg)
	
	assert.Equal(t, 2, state.NumSubscribers(), "should have 2 subscribers")
}

type MockProcessor struct {
	receivedMessages []actor.Message
}

func (m *MockProcessor) Process(msg actor.Message) {
	m.receivedMessages = append(m.receivedMessages, msg)
}

func (m *MockProcessor) Shutdown() {}

func (m *MockProcessor) GetState() any {
	return m.receivedMessages
}

func TestNotifySubscribersWithDelivery(t *testing.T) {
	mockProcessor1 := &MockProcessor{}
	mockProcessor2 := &MockProcessor{}
	
	subAddr1 := actor.NewAddress("local", "subscriber1")
	subAddr2 := actor.NewAddress("local", "subscriber2")
	
	_, err := actor.RegisterActor(subAddr1, mockProcessor1)
	assert.NoError(t, err)
	defer actor.UnRegisterActor(subAddr1)
	
	_, err = actor.RegisterActor(subAddr2, mockProcessor2)
	assert.NoError(t, err)
	defer actor.UnRegisterActor(subAddr2)
	
	state := subscriber.NewSubscriptionsState()
	state.AddSubscription(subAddr1)
	state.AddSubscription(subAddr2)
	
	fromAddr := actor.NewAddress("local", "sender")
	msg := actor.NewMessage(nil, fromAddr, "test message")
	
	state.NotifySubscribers(msg)
	
	assert.Eventually(t, func() bool {
		return len(mockProcessor1.receivedMessages) == 1 && len(mockProcessor2.receivedMessages) == 1
	}, 100*time.Millisecond, 10*time.Millisecond, "both subscribers should receive the message")
	
	assert.Equal(t, subAddr1, mockProcessor1.receivedMessages[0].To, "message should be addressed to subscriber1")
	assert.Equal(t, subAddr2, mockProcessor2.receivedMessages[0].To, "message should be addressed to subscriber2")
	assert.Equal(t, "test message", mockProcessor1.receivedMessages[0].Body, "message body should match")
	assert.Equal(t, "test message", mockProcessor2.receivedMessages[0].Body, "message body should match")
}
