package batch_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/cinecity/pkg/batch"
	"github.com/stretchr/testify/assert"
)

// type mockProcessor struct {
// 	state    string
// 	messages []actor.Message
// }

// func (m *mockProcessor) Process(msg actor.Message) {
// 	m.messages = append(m.messages, msg)
// 	switch payload := msg.Body.(type) {
// 	case FirstMessage:
// 		m.state = fmt.Sprintf("write by batcher: %s", payload)
// 	case SecondMessage:
// 		m.state = fmt.Sprintf("write by batcher: %s", payload)
// 	}
// }

// func (m *mockProcessor) Shutdown() {
// 	m.messages = nil
// 	m.state = ""
// }

// func (m *mockProcessor) GetState() any {
// 	return m.state
// }

// func newMockProcessor() *mockProcessor {
// 	processor := &mockProcessor{
// 		state:    "initial",
// 		messages: make([]actor.Message, 0),
// 	}
// 	return processor
// }

type FirstMessage string
type SecondMessage string

var fixtureState = make([]actor.Message, 0)

func setup() (
	msg1 actor.Message,
	msg2 actor.Message,
	// actor1 *actor.Actor,
	// actor2 *actor.Actor,
	handler func(msg actor.Message),
) {
	fromAddr := actor.NewAddress("lcl", "Actor-A-")
	toAddr := actor.NewAddress("lcl", "Actor-B-")

	var msg1Body FirstMessage = "hello"
	msg1 = actor.NewMessage(
		fromAddr,
		toAddr,
		msg1Body,
	)

	var msg2Body SecondMessage = "world"
	msg2 = actor.NewMessage(
		fromAddr,
		toAddr,
		msg2Body,
	)

	fixtureState = make([]actor.Message, 0)

	// sp := newMockProcessor()
	// actor1, _ = actor.RegisterActor(fromAddr, sp)
	// actor2, _ = actor.RegisterActor(toAddr, sp)
	handler = func(msg actor.Message) {
		fixtureState = append(fixtureState, msg)
		slog.Info("write by batcher")
		// switch payload := msg.Body.(type) {
		// case FirstMessage:
		// 	fmt.Println("write by batcher: ", payload)
		// case SecondMessage:
		// 	fmt.Println("write by batcher: ", payload)
		// }
	}
	return
}

func TestBatcher_StateChangesAfterTimeoutLimits(t *testing.T) {
	msg1, msg2, handler := setup()

	b := batch.NewBatcher(1000, 100, handler)
	b.Add(msg1)
	b.Add(msg2)

	<-time.After(1010 * time.Millisecond)
	assert.Equal(t, 2, len(fixtureState), "It must be persisted 2 messages")
	assert.IsType(t, FirstMessage(""), fixtureState[0].Body, "It must be persisted the first message sended")
	assert.IsType(t, SecondMessage(""), fixtureState[1].Body, "It must be persisted the second message sended")
}
func TestBatcher_StateNotChangesAfterTimeoutLimits(t *testing.T) {
	msg1, msg2, handler := setup()

	b := batch.NewBatcher(1000, 100, handler)
	b.Add(msg1)
	b.Add(msg2)

	<-time.After(500 * time.Millisecond)
	assert.Equal(t, 0, len(fixtureState), "It must be persisted 0 messages")
	<-time.After(510 * time.Millisecond)
	assert.Equal(t, 2, len(fixtureState), "It must be persisted 2 messages")
}

func TestBatcher_StateChangesByMaxItemsLimits(t *testing.T) {
	msg1, msg2, handler := setup()

	b := batch.NewBatcher(1000, 2, handler)
	b.Add(msg1)
	b.Add(msg2)

	<-time.After(100 * time.Millisecond)
	assert.Equal(t, 2, len(fixtureState), "It must be persisted 2 messages")
	assert.IsType(t, FirstMessage(""), fixtureState[0].Body, "It must be persisted the first message sended")
	assert.IsType(t, SecondMessage(""), fixtureState[1].Body, "It must be persisted the second message sended")
}

func TestBatcher_StateDontChangesBeforMaxItemsLimits(t *testing.T) {
	msg1, msg2, handler := setup()
	b := batch.NewBatcher(1000, 3, handler)
	b.Add(msg1)
	b.Add(msg2)

	<-time.After(500 * time.Millisecond)
	assert.Equal(t, 0, len(fixtureState), "It must be persisted 0 messages")

	<-time.After(550 * time.Millisecond)
	assert.Equal(t, 2, len(fixtureState), "It must be persisted 2 messages")
	assert.IsType(t, FirstMessage(""), fixtureState[0].Body, "It must be persisted the first message sended")
	assert.IsType(t, SecondMessage(""), fixtureState[1].Body, "It must be persisted the second message sended")
}
