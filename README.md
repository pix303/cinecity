# Cinecity

A simple and lightweight implementation of the Actor pattern in Go, designed to better understand the pattern, tested to verify its effectiveness in a personal project [https://github.com/pix303/localemgmt-go](localemgmt-go).
For other good examples of actor model implementation in Go check out [go-actor](https://github.com/vladopajic/go-actor) or [Hollywood](https://github.com/anthdm/hollywood) and similar to actor model but very complete and extended [Watermill](https://github.com/ThreeDotsLabs/watermill)

## Overview

The Actor model promises to simplify and streamline the build process and interactions between application components, providing a model that reflects reality. This is achieved through the exchange of messages within actors. Each actor is responsible for updating its own state.

An Actor is composed of:
- An address defined by an area and an id
- A state defined and modified by a StateProcessor interface

An **Actor** can:
- Send messages to other actors
- Change their own state in response to messages
- Subscribe to future message of an other actor
- Create new actors

**Postman** is a manager entity in this model and it is responsible to inbox messages ("send and forget" approach) and to inbox and wait a response ("ask" approach).

## How to process message

The core method to implement StateProcessor interface is **Process** that takes a message as parameter, evaluate the body type and act consequentially 

```go
...
func (state *ProductsState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {

	case AddNewProductPayload:
		slog.Info("handling add product message")
		p := payload.Product
		state.products = append(state.products, p)

	case AddQuantityToProductPayload:
		slog.Info("handling add quantity to product message")
		p := state.getProduct(payload.Code)
		if p != nil {
			p.Quantity += payload.Quantity
			slog.Info("update quantity with success", slog.Int("qty", p.Quantity))
		} else {
			slog.Info("update quantity fail")
		}
	...
```

## Init an actor
Before to start using an actor it needs to create and register it

```go
state := NewProductState() 
warehouseAddress := actor.NewAddress("local", "warehouse")
warehouseActor, err := actor.RegisterActor(
  warehouseAddress,
	state, 
)
```

## Send messages
Messages are shipping in an async mode; here an example to create a message and just **send it and forget**

```go
msg := actor.NewMessage(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Code: "ABC", Quantity: 2}, // body
	)

err := actor.SendMessage(msg)
	
```

Here an example to **send a message and wait a response**

```go
msg := actor.NewMessageWithResponse(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Code: "ABC", Quantity: 2}, // body
	)

response, err := actor.SendMessageWithResponse[AddQuantityToProductResultPayload](msg)
	
```

A message can be sent to a group of actors organized by address area:

```go
msg := actor.NewMessage(
		warehouseAddress,
		customerAddress,
		NewProductAdded{Code: "ABC"},
	)

// address area filter is optional
filterByArea := "products"
numActorsReached := actor.BroadcastMessage(msg, &filterByArea)
	
```

## Subscribe to receive messages
An actor can subscribe to messages sent by another actor. The notifying actor will determine how to implement in the Process function which messages must be notified

```go
type NotifierActorProcessor struct {
	state    StateType
	notifier *subscriber.SubscriptionsState
}


func (a *NotifierActorProcessor ) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	// library message body to add a subscription
	case actor.AddSubscriptionMessageBody:
		a.notifier.AddSubscription(msg.From)
	// library message body to remove a subscription
	case actor.RemoveSubscriptionMessageBody :
		a.notifier.RemoveSubscription(msg.From)
	// example of a message that trigger notify the subscribers	
	case TriggerSubscriptionNotifierBodyMsg:
		subsMsg := actor.NewSubscribersMessage(msg.To, "hello subscribers!")
		m.notifier.NotifySubscribers(subsMsg)
	
```

## Debounce and process messages in batch
Messages can be batched together to avoid unnecessary processing of single messages

```go
type State struct {
	state     StateType
	batcher   *batch.Batcher
}

func NewState() *State{
	...
	// wait max 5 seconds or max 5 messages, than trigger the handler function
	b := batch.NewBatcher(5000, 5, state.updateItem)
	s.batcher = b
	...
}

func (state *State) Process(msg actor.Message) {
	switch msg.Body.(type) {
	// enqueue messages
	case ItemUpdateMsgPayload:
		state.batcher.Add(msg)
	}
}

// take only one message at a time; the batcher will loop through all the messages
func (s State) updateItem(msg actor.Message){
	// update state
}
```
