package actor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/cinecity/pkg/subscriber"
	"github.com/stretchr/testify/assert"
)

type mockProcessor struct {
	state    string
	messages []actor.Message
	notifier *subscriber.SubscriptionsState
}

type TriggerSubscriptionNotifierBodyMsg string
type WithReturnTriggerMsgBody struct {
	Content string
}
type WithReturnTriggerMsgBodyReturn string

func (m *mockProcessor) Process(msg actor.Message) {
	m.messages = append(m.messages, msg)
	switch payload := msg.Body.(type) {
	case actor.AddSubscriptionMessageBody:
		m.notifier.AddSubscription(msg.From)
	case TriggerSubscriptionNotifierBodyMsg:
		subsMsg := actor.NewSubscribersMessage(msg.To, "hello subscribers!")
		m.notifier.NotifySubscribers(subsMsg)
	case WithReturnTriggerMsgBody:
		rBody := WithReturnTriggerMsgBodyReturn(fmt.Sprintf("returned: %s", payload.Content))
		if msg.WithResponse {
			returnMsg := actor.NewReturnMessage(rBody, msg, nil)
			msg.ResponseChan <- returnMsg
		}
	case string:
		m.state = msg.Body.(string)
	}
}

func (m *mockProcessor) Shutdown() {
	m.messages = nil
	m.state = ""
}

func (m *mockProcessor) GetState() any {
	return m.state
}

func newMockProcessor() *mockProcessor {
	processor := &mockProcessor{
		state:    "initial",
		messages: make([]actor.Message, 0),
		notifier: subscriber.NewSubscriptionsState(),
	}
	return processor
}

func TestGetPostman(t *testing.T) {
	postman1 := actor.GetPostman()
	postman2 := actor.GetPostman()

	assert.Equal(t, postman1, postman2, "GetPostman() should return the same instance (singleton)")
	assert.NotNil(t, postman1.GetContext(), "Postman context should not be nil")
}

func TestRegisterActor(t *testing.T) {
	address := actor.NewAddress("test", "actor1")
	processor := newMockProcessor()
	registeredActor, err := actor.RegisterActor(address, processor)
	assert.NoError(t, err, "Failed to register actor")
	assert.NotNil(t, registeredActor, "Registered actor should not be nil")
	assert.Equal(t, address.String(), registeredActor.GetAddress().String(), "Address mismatch")
	assert.Equal(t, 1, actor.NumActors(), "Expected 1 actor")
}

func TestRegisterActorInvalidAddress(t *testing.T) {
	processor := newMockProcessor()

	tests := []struct {
		name    string
		address *actor.Address
	}{
		{"nil address", nil},
		{"empty area", actor.NewAddress("", "id")},
		{"empty id", actor.NewAddress("area", "")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := actor.RegisterActor(tt.address, processor)
			assert.Error(t, err, "Expected error for invalid address")
			assert.Equal(t, actor.ErrAddressInvalid, err, "Expected ErrAddressInvalid")
		})
	}
}

func TestRegisterActorDuplicate(t *testing.T) {
	address := actor.NewAddress("test", "duplicate")
	processor1 := newMockProcessor()
	processor2 := newMockProcessor()

	_, err := actor.RegisterActor(address, processor1)
	assert.NoError(t, err, "Failed to register first actor")

	_, err = actor.RegisterActor(address, processor2)
	assert.Error(t, err, "Expected error when registering duplicate actor")
	assert.Equal(t, actor.ErrActorAddressAlreadyRegistered, err, "Expected ErrActorAddressAlreadyRegistered")
}

func TestUnRegisterActor(t *testing.T) {
	actor.ShutdownAll()
	address := actor.NewAddress("test", "unregister")
	processor := newMockProcessor()

	actor.RegisterActor(address, processor)
	assert.Equal(t, 1, actor.NumActors(), "Expected 1 actor before unregister")

	actor.UnRegisterActor(address)
	assert.Equal(t, 0, actor.NumActors(), "Expected 0 actors after unregister")
}

func TestSendMessage(t *testing.T) {
	actor.ShutdownAll()
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	processor := newMockProcessor()

	_, err := actor.RegisterActor(toAddr, processor)
	assert.NoError(t, err, "Failed to register receiver")

	msg := actor.NewMessage(toAddr, fromAddr, "test message")
	err = actor.SendMessage(msg)
	assert.NoError(t, err, "Failed to send message")

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, len(processor.messages), "Expected 1 message")
	assert.Equal(t, "test message", processor.messages[0].Body, "Message body mismatch")
}

func TestSendMessageToNonExistentActor(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "nonexistent")

	msg := actor.NewMessage(toAddr, fromAddr, "test message")
	err := actor.SendMessage(msg)
	assert.Error(t, err, "Expected error when sending to non-existent actor")
	assert.Equal(t, actor.ErrActorNotFound, err, "Expected ErrActorNotFound")
}

func TestSendMessageWithResponse(t *testing.T) {
	actor.ShutdownAll()
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")

	processor := newMockProcessor()
	_, err := actor.RegisterActor(toAddr, processor)
	assert.NoError(t, err, "Failed to register receiver")
	msgBody := WithReturnTriggerMsgBody{
		Content: "test message",
	}
	msg := actor.NewMessageWithResponse(toAddr, fromAddr, msgBody)
	response, err := actor.SendMessageWithResponse[WithReturnTriggerMsgBodyReturn](msg)
	<-time.After(200 * time.Millisecond)
	assert.NoError(t, err, "Failed to send message with response")
	assert.IsType(t, WithReturnTriggerMsgBodyReturn(""), response, "Expected response")
	assert.Contains(t, fmt.Sprintf("returned: %s", msgBody.Content), response, "Expected response as trigger message prefixed by 'returned' string")
}

func TestSendMessageWithResponseToNonExistentActor(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "nonexistent")

	msg := actor.NewMessageWithResponse(toAddr, fromAddr, "test message")
	_, err := actor.SendMessageWithResponse[string](msg)
	assert.Error(t, err, "Expected error when sending to non-existent actor")
	assert.Equal(t, actor.ErrActorNotFound, err, "Expected ErrActorNotFound")
}

func TestBroadcastMessageToAll(t *testing.T) {
	actor.ShutdownAll()
	fromAddr := actor.NewAddress("test", "sender")
	toAddr1 := actor.NewAddress("test", "receiver1")
	toAddr2 := actor.NewAddress("test", "receiver2")

	processor1 := newMockProcessor()
	processor2 := newMockProcessor()

	actor.RegisterActor(toAddr1, processor1)
	actor.RegisterActor(toAddr2, processor2)

	msg := actor.NewMessage(toAddr1, fromAddr, "broadcast message")
	numSent := actor.BroadcastMessage(msg, nil)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, len(processor1.messages), "Expected 1 message for receiver1")
	assert.Equal(t, 1, len(processor2.messages), "Expected 1 message for receiver2")
	assert.Equal(t, 2, numSent, "Expected 2 messages to be sent")
}

func TestBroadcastMessageByArea(t *testing.T) {
	actor.ShutdownAll()
	fromAddr := actor.NewAddress("local", "sender")
	toLocalAddr := actor.NewAddress("local", "receiver")
	toRemoteAddr := actor.NewAddress("remote", "receiver")

	processorLocal := newMockProcessor()
	processorRemote := newMockProcessor()

	actor.RegisterActor(toLocalAddr, processorLocal)
	actor.RegisterActor(toRemoteAddr, processorRemote)

	msg := actor.NewMessage(toLocalAddr, fromAddr, "broadcast message")
	area := "local"
	numSent := actor.BroadcastMessage(msg, &area)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, len(processorLocal.messages), "Expected 1 message for local receiver")
	assert.Equal(t, 0, len(processorRemote.messages), "Expected 0 message for remote receiver")
	assert.Equal(t, 1, numSent, "Expected 1 messages to be sent")
}

func TestSubscrption(t *testing.T) {
	actor.ShutdownAll()
	fromAddr := actor.NewAddress("test", "subscriber")
	toAddr := actor.NewAddress("test", "notifier")

	processor := newMockProcessor()
	processorSub := newMockProcessor()
	actor.RegisterActor(toAddr, processor)
	actor.RegisterActor(fromAddr, processorSub)

	subMsg := actor.NewAddSubcriptionMessage(fromAddr, toAddr)
	actor.SendMessage(subMsg)
	var triggerMsgBody TriggerSubscriptionNotifierBodyMsg = "spread your state"
	triggerMsg := actor.NewMessage(toAddr, fromAddr, triggerMsgBody)
	actor.SendMessage(triggerMsg)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 2, len(processor.messages), "Expected 2 subscription message")
	assert.Equal(t, 1, len(processorSub.messages), "Expected 1 subscription message")
	assert.Contains(t, processorSub.GetState(), "hello", "Expected state modified by from reciver message")
}

func TestNumActors(t *testing.T) {
	actor.ShutdownAll()
	initialCount := actor.NumActors()

	address1 := actor.NewAddress("test", "actor1")
	address2 := actor.NewAddress("test", "actor2")
	processor := &mockProcessor{state: "initial"}

	actor.RegisterActor(address1, processor)
	assert.Equal(t, initialCount+1, actor.NumActors(), "Actor count mismatch after first registration")

	actor.RegisterActor(address2, processor)
	assert.Equal(t, initialCount+2, actor.NumActors(), "Actor count mismatch after second registration")

	actor.RegisterActor(address1, processor)
	actor.RegisterActor(address2, processor)
	assert.Equal(t, initialCount+2, actor.NumActors(), "Actor count mismatch after others failing registrations")

	actor.UnRegisterActor(address1)
	assert.Equal(t, initialCount+1, actor.NumActors(), "Actor count mismatch after unregister")
}

func TestShutdownAll(t *testing.T) {
	actor.ShutdownAll()
	address1 := actor.NewAddress("test", "shutdown1")
	address2 := actor.NewAddress("test", "shutdown2")
	processor := &mockProcessor{state: "initial"}

	actor.RegisterActor(address1, processor)
	actor.RegisterActor(address2, processor)

	assert.Greater(t, actor.NumActors(), 0, "Expected actors to be registered before shutdown")

	actor.ShutdownAll()

	assert.Equal(t, 0, actor.NumActors(), "Expected 0 actors after shutdown")

	ctx := actor.GetPostman().GetContext()
	select {
	case <-ctx.Done():
	default:
		t.Error("Expected context to be cancelled after shutdown")
	}
}
