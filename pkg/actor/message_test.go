package actor_test

import (
	"errors"
	"testing"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	body := "test message"

	msg := actor.NewMessage(toAddr, fromAddr, body)

	assert.Equal(t, toAddr, msg.To, "To address should match")
	assert.Equal(t, fromAddr, msg.From, "From address should match")
	assert.Equal(t, body, msg.Body, "Body should match")
	assert.False(t, msg.WithReturn, "WithReturn should be false")
	assert.Nil(t, msg.ReturnChan, "ReturnChan should be nil")
	assert.Equal(t, 0, msg.ReturnTimeout, "ReturnTimeout should be 0")
}

func TestNewMessageWithReturn(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	body := "test message"

	msg := actor.NewMessageWithResponse(toAddr, fromAddr, body)

	assert.Equal(t, toAddr, msg.To, "To address should match")
	assert.Equal(t, fromAddr, msg.From, "From address should match")
	assert.Equal(t, body, msg.Body, "Body should match")
	assert.True(t, msg.WithReturn, "WithReturn should be true")
	assert.NotNil(t, msg.ReturnChan, "ReturnChan should not be nil")
	assert.Equal(t, 60, msg.ReturnTimeout, "ReturnTimeout should be 60")
}

func TestNewReturnMessage(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	originalBody := "original message"
	returnBody := "return message"

	originalMsg := actor.NewMessage(toAddr, fromAddr, originalBody)
	returnMsg := actor.NewReturnMessage(returnBody, originalMsg, errors.New("test error"))

	assert.Equal(t, fromAddr, returnMsg.Message.To, "Return message To should be original From")
	assert.Equal(t, toAddr, returnMsg.Message.From, "Return message From should be original To")
	assert.Equal(t, returnBody, returnMsg.Message.Body, "Return message body should match")
	assert.False(t, returnMsg.Message.WithReturn, "Return message should not have return")
	assert.Nil(t, returnMsg.Message.ReturnChan, "Return message ReturnChan should be nil")
	assert.NotNil(t, returnMsg.Err, "Return message should be contain an error")
	assert.Contains(t, "test error", returnMsg.Err.Error(), "Return error message should be consistent")
}

func TestNewReturnMessageWithNullError(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	originalBody := "original message"
	returnBody := "return message"

	originalMsg := actor.NewMessage(toAddr, fromAddr, originalBody)
	returnMsg := actor.NewReturnMessage(returnBody, originalMsg, nil)

	assert.Equal(t, fromAddr, returnMsg.Message.To, "Return message To should be original From")
	assert.Equal(t, toAddr, returnMsg.Message.From, "Return message From should be original To")
	assert.Equal(t, returnBody, returnMsg.Message.Body, "Return message body should match")
	assert.False(t, returnMsg.Message.WithReturn, "Return message should not have return")
	assert.Nil(t, returnMsg.Message.ReturnChan, "Return message ReturnChan should be nil")
	assert.Nil(t, returnMsg.Err, "Return message should not contain an error")
}

func TestMessageSetTimeout(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	body := "test message"

	msg := actor.NewMessage(toAddr, fromAddr, body)
	assert.Equal(t, 0, msg.ReturnTimeout, "Initial timeout should be 0")

	msg.SetTimeout(30)
	assert.Equal(t, 30, msg.ReturnTimeout, "Timeout should be updated to 30")

	msg.SetTimeout(120)
	assert.Equal(t, 120, msg.ReturnTimeout, "Timeout should be updated to 120")
}

func TestNewAddSubscriptionMessage(t *testing.T) {
	subscriberAddr := actor.NewAddress("test", "subscriber")
	notifierAddr := actor.NewAddress("test", "notifier")

	msg := actor.NewAddSubcriptionMessage(subscriberAddr, notifierAddr)

	assert.Equal(t, subscriberAddr, msg.From, "From should be subscriber address")
	assert.Equal(t, notifierAddr, msg.To, "To should be notifier address")
	assert.IsType(t, actor.AddSubscriptionMessageBody{}, msg.Body, "Body should be AddSubscriptionMessageBody")
}

func TestNewRemoveSubscriptionMessage(t *testing.T) {
	subscriberAddr := actor.NewAddress("test", "subscriber")
	notifierAddr := actor.NewAddress("test", "notifier")

	msg := actor.NewRemoveSubscriptionMessage(subscriberAddr, notifierAddr)

	assert.Equal(t, subscriberAddr, msg.From, "From should be subscriber address")
	assert.Equal(t, notifierAddr, msg.To, "To should be notifier address")
	assert.IsType(t, actor.RemoveSubscriptionMessageBody{}, msg.Body, "Body should be RemoveSubscriptionMessageBody")
}

func TestNewSubscribersMessage(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	body := []string{"subscriber1", "subscriber2"}

	msg := actor.NewSubscribersMessage(fromAddr, body)

	assert.Equal(t, fromAddr, msg.From, "From should match")
	assert.Equal(t, body, msg.Body, "Body should match")
	assert.Nil(t, msg.To, "To should be nil for subscribers message")
}

func TestMessageString(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	body := "test message"

	msg := actor.NewMessage(toAddr, fromAddr, body)
	expected := "from: test.sender to: test.receiver with body: test message"

	assert.Equal(t, expected, msg.String(), "String representation should match expected format")
}

func TestMessageStringWithNilAddresses(t *testing.T) {
	body := "test message"

	msg := actor.Message{
		From: nil,
		To:   nil,
		Body: body,
	}

	expected := "from: address nil to: address nil with body: test message"
	assert.Equal(t, expected, msg.String(), "String representation should handle nil addresses")
}

func TestMessageStringWithComplexBody(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	body := map[string]interface{}{"key": "value", "number": 42}

	msg := actor.NewMessage(toAddr, fromAddr, body)
	str := msg.String()

	assert.Contains(t, str, "from: test.sender", "String should contain from address")
	assert.Contains(t, str, "to: test.receiver", "String should contain to address")
	assert.Contains(t, str, "with body:", "String should contain body prefix")
}

func TestEmptyMessage(t *testing.T) {
	assert.Equal(t, actor.Message{}, actor.EmptyMessage, "EmptyMessage should be empty Message struct")
}

func TestMessageReturnChannel(t *testing.T) {
	fromAddr := actor.NewAddress("test", "sender")
	toAddr := actor.NewAddress("test", "receiver")
	body := "test message"

	msg := actor.NewMessageWithResponse(toAddr, fromAddr, body)

	// Test that ReturnChan is a buffered channel with capacity 1
	assert.Equal(t, 1, cap(msg.ReturnChan), "ReturnChan should have capacity 1")

	// Test sending and receiving from ReturnChan
	wrappedMsg := actor.NewReturnMessage("return content", msg, nil)
	select {
	case msg.ReturnChan <- wrappedMsg:
		assert.True(t, true, "Should be able to send to ReturnChan")
	default:
		t.Error("Should be able to send to ReturnChan")
	}

	select {
	case received := <-msg.ReturnChan:
		assert.Equal(t, wrappedMsg, received, "Should receive the same wrapped message")
	default:
		t.Error("Should be able to receive from ReturnChan")
	}
}
