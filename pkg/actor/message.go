package actor

import "fmt"

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

type AddSubscriptionMessageBody struct{}

func NewAddSubcriptionMessage(subscriberAddress *Address, notifierAddress *Address) Message {
	return Message{
		From: subscriberAddress,
		To:   notifierAddress,
		Body: AddSubscriptionMessageBody{},
	}
}

type RemoveSubscriptionMessageBody struct{}

func NewRemoveSubscriptionMessage(subscriberAddress *Address, notifierAddress *Address) Message {
	return Message{
		From: subscriberAddress,
		To:   notifierAddress,
		Body: RemoveSubscriptionMessageBody{},
	}
}

func NewSubscribersMessage(from *Address, body any) Message {
	return Message{
		From: from,
		Body: body,
	}
}

func (this *Message) String() string {
	return fmt.Sprintf("from: %s to: %s with body: %v", this.From.String(), this.To.String(), this.Body)
}
