# Cinecity

A simple and lightweight implementation of the Actor pattern in Go, designed to better understand the pattern, tested to verify its effectiveness in a personal project [https://github.com/pix303/localemgmt-go](localemgmt-go).
For examples of actor model implementation in Go you should see [go-actor](https://github.com/vladopajic/go-actor) or [hollywood](https://github.com/anthdm/hollywood)

## Overview

The Actor model promises to simplify and streamline the build process and interactions between application components, providing a model that reflects reality. This is achieved through the exchange of messages within actors. Each actor is responsible for updating its own state.

An **Actor** can:
- Send messages to other actors
- Change their own state in response to messages
- Subscribe to future message of an actor
- Create new actors

An Actor is composed of:
- an address defined by an area and an id
- a state defined by StateProcessor interface

**Postman** is responsible to inbox messages (send message and forget approach) and also to inbox and wait a response (ask approach). It also can send message outside an actor.

## Examples

The core method to implement in a state processor is **Process** that takes a message as parameter, evaluate the body type and act consequentially 

```go
...
func (state *ProductsState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {

	case AddNewProductPayload:
		slog.Info("AddProductMsg")
		p := payload.Product
		state.products = append(state.products, p)

	case AddQuantityToProductPayload:
		slog.Info("AddQuantityProductMsg")
		p := payload.Product
		cp := state.getProduct(p.Code)
...
```

Before to start using an actor it needs to create and register it

```go

warehouseAddress := actor.NewAddress("local", "warehouse")
warehouseActor, err := actor.RegisterActor(
  warehouseAddress,
	// usually create state processor before to handle errors
	NewProductState(), 
)
```

Messages are shipping in a async mode; here an example to create a message and just **send it and forget**

```go
msg := actor.NewMessage(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Product{Code: "ABC", Quantity: 2}}, // body
	)

err := actor.SendMessage(msg)
	
```

Here an example to **send a message and wait a response**

```go
msg := actor.NewMessageWithResponse(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Product{Code: "ABC", Quantity: 2}}, // body
	)

response, err := actor.SendMessageWithResponse[AddQuantityToProductResultPayload](msg)
	
```

A message can be broadcast to a group of actors

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

An actor can subscribe to messages sent by another actor. The notifying actor will determine how to implement the Process function to decide which messages will be sent

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

Messages can be batched together to avoid unnecessary processing of single messages

```go
type State struct {
	state     StateType
	batcher   *batch.Batcher
}

func NewState() *State{
	...
	b := batch.NewBatcher(5000, 5, state.updateItem)
	s.batcher = b
	...
}

func (state *State) Process(msg actor.Message) {
	switch msg.Body.(type) {
	case ItemUpdateMsgPayload:
		state.batcher.Add(msg)
	}
}
```
