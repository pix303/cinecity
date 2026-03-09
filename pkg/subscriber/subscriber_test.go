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
	subsActor := subscriber.NewSubscription()
	assert.Equal(t, 0, subsActor.NumSubscribers(), "initial num of subscribers must be 0")
}

func TestNewSubscribersMessage(t *testing.T) {
	fromAddr := actor.NewAddress("local", "sender")
	body := "test body"

	msg := subscriber.NewSubscribersMessage(fromAddr, body)

	assert.Equal(t, fromAddr, msg.From, "From should match")
	assert.Equal(t, body, msg.Body, "Body should match")
	assert.Nil(t, msg.To, "To should be nil")
	assert.False(t, msg.WithResponse, "WithResponse should be false")
}

func TestAddSubscriber(t *testing.T) {
	subsActor := subscriber.NewSubscription()
	subAddr := actor.NewAddress("local", "subscriber")
	addMsg := subscriber.NewAddSubcriptionMessage(subAddr, nil)
	subsActor.Process(addMsg)
	assert.Equal(t, 1, subsActor.NumSubscribers(), "num of subscribers after add must be 1")
}

func TestRemoveSubscriber(t *testing.T) {
	subsActor := subscriber.NewSubscription()
	subAddr := actor.NewAddress("local", "subscriber")
	addMsg := subscriber.NewAddSubcriptionMessage(subAddr, nil)
	subsActor.Process(addMsg)
	removeMsg := subscriber.NewRemoveSubscriptionMessage(subAddr, nil)
	subsActor.Process(removeMsg)
	assert.Equal(t, 0, subsActor.NumSubscribers(), "num of subscribers after add and remove must be 0")
}

func TestNotifySubscribers(t *testing.T) {
	actor.InitPostman()
	subsActor := subscriber.NewSubscription()

	subAddr1 := actor.NewAddress("local", "subscriber1")
	subAddr2 := actor.NewAddress("local", "subscriber2")

	addMsg1 := subscriber.NewAddSubcriptionMessage(subAddr1, nil)
	addMsg2 := subscriber.NewAddSubcriptionMessage(subAddr2, nil)
	subsActor.Process(addMsg1)
	subsActor.Process(addMsg2)

	fromAddr := actor.NewAddress("local", "sender")
	msg := actor.NewMessage(nil, fromAddr, "test message")

	result := subsActor.NotifySubscribers(msg)

	assert.Equal(t, 2, result, "should have 2 subscribers")
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

	actor.InitPostman()
	_, err := actor.RegisterActor(subAddr1, mockProcessor1)
	assert.NoError(t, err)
	defer actor.UnRegisterActor(subAddr1)

	_, err = actor.RegisterActor(subAddr2, mockProcessor2)
	assert.NoError(t, err)
	defer actor.UnRegisterActor(subAddr2)

	subsActor := subscriber.NewSubscription()
	addMsg1 := subscriber.NewAddSubcriptionMessage(subAddr1, nil)
	addMsg2 := subscriber.NewAddSubcriptionMessage(subAddr2, nil)
	subsActor.Process(addMsg1)
	subsActor.Process(addMsg2)

	fromAddr := actor.NewAddress("local", "sender")
	msg := actor.NewMessage(nil, fromAddr, "test message")

	subsActor.NotifySubscribers(msg)

	assert.Eventually(t, func() bool {
		return len(mockProcessor1.receivedMessages) == 1 && len(mockProcessor2.receivedMessages) == 1
	}, 100*time.Millisecond, 10*time.Millisecond, "both subscribers should receive the message")

	assert.Equal(t, subAddr1, mockProcessor1.receivedMessages[0].To, "message should be addressed to subscriber1")
	assert.Equal(t, subAddr2, mockProcessor2.receivedMessages[0].To, "message should be addressed to subscriber2")
	assert.Equal(t, "test message", mockProcessor1.receivedMessages[0].Body, "message body should match")
	assert.Equal(t, "test message", mockProcessor2.receivedMessages[0].Body, "message body should match")
}
