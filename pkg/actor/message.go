package actor

import (
	"fmt"
)

type Message struct {
	From            *Address
	To              *Address
	Body            any
	WithResponse    bool
	ResponseChan    chan WrappedMessageWithError
	ResponseTimeout int
}

var EmptyMessage = Message{}

func NewMessage(to *Address, from *Address, body any) Message {
	return Message{
		To:              to,
		From:            from,
		Body:            body,
		WithResponse:    false,
		ResponseChan:    nil,
		ResponseTimeout: 0,
	}
}

func NewBroadcastMessage(from *Address, body any) Message {
	return Message{
		To:              nil,
		From:            from,
		Body:            body,
		WithResponse:    false,
		ResponseChan:    nil,
		ResponseTimeout: 0,
	}
}

func NewMessageWithResponse(to *Address, from *Address, body any) Message {
	c := make(chan WrappedMessageWithError, 1)
	return Message{
		To:              to,
		From:            from,
		Body:            body,
		WithResponse:    true,
		ResponseChan:    c,
		ResponseTimeout: 60,
	}
}

func (msg *Message) SetTimeout(value int) {
	msg.ResponseTimeout = value
}

type WrappedMessageWithError struct {
	Message *Message
	Err     error
}

func NewReturnMessage(body any, originalMessage Message, err error) WrappedMessageWithError {
	m := NewMessage(
		originalMessage.From,
		originalMessage.To,
		body,
	)
	return WrappedMessageWithError{&m, err}
}

func (msg *Message) String() string {
	return fmt.Sprintf("from: %s to: %s with body: %v", msg.From.String(), msg.To.String(), msg.Body)
}
