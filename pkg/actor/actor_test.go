package actor_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/stretchr/testify/assert"
)

type TestProcessorState struct {
	Data *string
}

func NewTestProcessorState() TestProcessorState {
	initalContent := "intial content"
	return TestProcessorState{
		Data: &initalContent,
	}
}

type NotHandledMessage string
type FirstMessage string
type SecondMessage string
type ThirdMessage string
type TestReturnMessage string
type WithSyncResponse string
type Response string

func (state *TestProcessorState) GetState() any {
	return state.Data
}

func (state *TestProcessorState) Process(msg actor.Message) {
	slog.Info("processing msg", slog.String("msg", msg.String()))
	switch msg.Body.(type) {
	case FirstMessage:
		r := fmt.Sprintf("processed by first event: %s", msg.Body)
		state.Data = &r
	case SecondMessage:
		r := fmt.Sprintf("%s ++ processed by second event: %s", *state.Data, msg.Body)
		state.Data = &r
	case ThirdMessage:
		r := fmt.Sprintf("processed by third event: %s", msg.Body)
		state.Data = &r
		var returnMsg TestReturnMessage = "return msg triggerd by third message"
		if msg.WithResponse {
			msg.ResponseChan <- actor.NewReturnMessage(returnMsg, msg, nil)
		}
	case TestReturnMessage:
		r := fmt.Sprintf("processed with return event: %s", msg.Body)
		state.Data = &r
	}
}

func (state *TestProcessorState) Shutdown() {
	state.Data = nil
	slog.Info("all clean after shutdown")
}

func generateActorForTest(address *actor.Address) *actor.Actor {
	processor := TestProcessorState{}
	a, err := actor.RegisterActor(address, &processor)
	if err != nil {
		panic(err)
	}
	return a
}

func setup() (reciver *actor.Actor, sender *actor.Actor) {
	reciver = generateActorForTest(actor.NewAddress("lcl", "Actor-A-"))
	sender = generateActorForTest(actor.NewAddress("lcl", "Actor-B-"))
	return
}

func Test_ShouldActorsBeActivateAtRegistration(t *testing.T) {
	reciver, sender := setup()

	assert.False(t, reciver.IsClosed(), "Receiver should not be closed after activation")
	assert.False(t, sender.IsClosed(), "Sender should not be closed after activation")
}

func Test_ShouldActorsDeactivate(t *testing.T) {
	actor.ShutdownAll()
	reciver, sender := setup()
	reciver.Deactivate()
	sender.Deactivate()

	assert.True(t, reciver.IsClosed(), "Receiver should be closed after deactivation")
	assert.True(t, sender.IsClosed(), "Sender should be closed after deactivation")
	actor.ShutdownAll()
}

func Test_ShouldSendMessageAndChangeActorState(t *testing.T) {
	reciver, sender := setup()

	var msgBody FirstMessage = "one"
	msg := actor.NewMessage(reciver.GetAddress(), sender.GetAddress(), msgBody)
	actor.SendMessage(msg)
	<-time.After(time.Millisecond * 10)
	state := reciver.GetState().(*string)
	assert.Contains(t, *state, "first event", "State should contain 'first event'")
	assert.Contains(t, *state, "one", "State should contain 'one'")
	actor.ShutdownAll()
}

func Test_ShouldSendWrongMessageAndNotChangeActorState(t *testing.T) {
	reciver, sender := setup()

	var msgBody NotHandledMessage = "wrong message"
	msg := actor.NewMessage(reciver.GetAddress(), sender.GetAddress(), msgBody)
	actor.SendMessage(msg)
	<-time.After(time.Millisecond * 10)
	state := reciver.GetState().(*string)
	assert.Nil(t, state, "State should be nil when wrong message is sent")
	actor.ShutdownAll()
}

func Test_ShouldSendMessageAndReciveResponse(t *testing.T) {
	reciver, sender := setup()

	var msgBody ThirdMessage = "third message"
	msg := actor.NewMessageWithResponse(reciver.GetAddress(), sender.GetAddress(), msgBody)
	rmsg, err := actor.SendMessageWithResponse[TestReturnMessage](msg)
	<-time.After(time.Millisecond * 10)
	assert.Nil(t, err, "Error should be nil when sending message with response")
	assert.Equal(t, TestReturnMessage("return msg triggerd by third message"), rmsg, "Response message should match expected return message")
	actor.ShutdownAll()
}

func Test_ShouldSendMessageAndReciveTimeout(t *testing.T) {
	reciver, sender := setup()

	var msgBody NotHandledMessage = "fake message"
	msg := actor.NewMessageWithResponse(reciver.GetAddress(), sender.GetAddress(), msgBody)
	msg.SetTimeout(10)
	_, err := actor.SendMessageWithResponse[TestReturnMessage](msg)
	<-time.After(time.Millisecond * 10)
	assert.Error(t, err, "Error should be valid")
	assert.IsType(t, actor.ErrSendWithReturnTimeout, err, "Response error should match expected return error type")
	actor.ShutdownAll()
}

func Test_ActorString(t *testing.T) {
	address := actor.NewAddress("local", "test-actor")
	processor := &TestProcessorState{}

	act, err := actor.RegisterActor(address, processor)
	assert.NoError(t, err)
	defer actor.UnRegisterActor(address)

	expected := "address: local.test-actor - isClosed: false"
	assert.Equal(t, expected, act.String(), "Actor string representation should match expected format")

	act.Deactivate()
	expectedClosed := "address: local.test-actor - isClosed: true"
	assert.Equal(t, expectedClosed, act.String(), "Actor string representation should reflect closed state")
}

func Test_ShouldReturnErrorForAClosedActor(t *testing.T) {
	address := actor.NewAddress("local", "test-actor")
	toAddr := actor.NewAddress("local", "test-actor")
	processor := &TestProcessorState{}

	act, err := actor.RegisterActor(address, processor)
	assert.NoError(t, err)
	defer actor.UnRegisterActor(address)

	act.Deactivate()
	msg := actor.NewMessage(toAddr, address, "hello")
	err = act.Inbox(msg)
	assert.Error(t, err, "Error should be returned for a closed actor")
	assert.IsType(t, actor.ErrInboxClosed, err, "Error should be of type ErrInboxClosed")
	_, err = act.InboxAndWaitResponse(msg)
	assert.Error(t, err, "Error should be returned for a closed actor")
	assert.IsType(t, actor.ErrInboxClosed, err, "Error should be of type ErrInboxClosed")
}
