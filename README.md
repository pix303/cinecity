# Cinecity

A simple and lightweight implementation of the Actor pattern in Go, designed to better understand the pattern, tested to verify its effectiveness in a personal project [https://github.com/pix303/localemgmt-go](localemgmt-go).
For other good examples of actor model implementation in Go check out [go-actor](https://github.com/vladopajic/go-actor) or [Hollywood](https://github.com/anthdm/hollywood) and similar to actor model but very complete and extended [Watermill](https://github.com/ThreeDotsLabs/watermill)

## Overview

The actor model promises to simplify and streamline the build process and interactions between application components, providing a model that reflects reality. This is achieved through the exchange of messages within actors. Each actor is responsible for updating its own state.

### Actor

An Actor is composed of:
- an address defined by an area and an id
- a state defined and modified by a StateProcessor interface

An Actor can:
- change its own state in response to messages
- send a message to another actor
- send a message to another actor and wait a response
- broadcast messages to subscribers
- subscribe to messages from another actor
- create new actors

### Message

A Message is the entity used by actors to exchange data. It's composed of:
- the destination address
- the sender address
- a body that represents the exchanged data
- a flag to set if a response is needed

### StateProcessor

The StateProcessor is an interface that represents the minimum API for handling the state managed by an actor. It takes care of processing incoming messages, cleanup before actor closure, and returning the current state.


### Postman
Postman is a manager responsible for register a new actor and his address, delivering messages ("send and forget" approach) and delivering and waiting for a response ("ask" approach).


## How to process a message

The core method to implement the StateProcessor interface is **Process** that takes a message as parameter, evaluates the body type, and acts accordingly. 

```go
...
func (state *ProductsState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {

	case AddNewProductPayload:
		p := payload.Product
		state.products = append(state.products, p)

	case AddQuantityToProductPayload:
		p := state.getProduct(payload.Code)
		if p != nil {
			p.Quantity += payload.Quantity
			slog.Info("update quantity with success", slog.Int("qty", p.Quantity))
		} else {
			slog.Error("failed to update product quantity: missing code", slog.String("code",p.Code))
		}
	...
```

## Initialize an actor
Before using an actor, you need to create and register it

```go
state := NewProductState() 
warehouseAddress := actor.NewAddress("main", "warehouse")
warehouseActor, err := actor.RegisterActor(
  warehouseAddress,
	state, 
)
```

## Send messages
Messages are sent asynchronously; here is an example to create a message and just **send it and forget**

```go
msg := actor.NewMessage(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Code: "ABC", Quantity: 2}, // body
	)

err := actor.SendMessage(msg)
	
```

Here is an example to **send a message and wait for a response**

```go
msg := actor.NewMessageWithResponse(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Code: "ABC", Quantity: 2}, // body
	)
// set the response type expected
response, err := actor.SendMessageWithResponse[AddQuantityToProductResultPayload](msg)
	
```

A message can be broadcasted to a group of actors organized by address area:

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
An actor can subscribe to messages sent by another actor. The notifying actor will determine which messages must be notified in the Process function

```go
type NotifierActorProcessor struct {
	state    StateType
	notifier *subscriber.SubscriptionsState
}


func (a *NotifierActorProcessor ) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	// library available message body type for adding a subscription
	case actor.AddSubscriptionMessageBody:
		a.notifier.AddSubscription(msg.From)
	// library available message body type for removing a subscription
	case actor.RemoveSubscriptionMessageBody:
		a.notifier.RemoveSubscription(msg.From)
	// example of a message that trigger notify the subscribers	
	case TriggerSubscriptionNotifierBodyMsg:
		subsMsg := actor.NewSubscribersMessage(msg.To, "hello subscribers!")
		m.notifier.NotifySubscribers(subsMsg)
	
```

## Batch messages
Messages can be batched together to avoid unnecessary processing of single messages.

```go
type State struct {
	state     StateType
	updateBatcher   *batch.Batcher
}

func NewState() *State{
	...
	// wait max 5 seconds or max 5 messages, than trigger the handler function
	b := batch.NewBatcher(5000, 5, state.updateItem)
	s.updateBatcher = b
	...
}

func (state *State) Process(msg actor.Message) {
	switch msg.Body.(type) {
	// enqueue messages
	case ItemUpdateMsgPayload:
		state.updateBatcher.Add(msg)
	}
}

// take only one message at a time; the batcher will loop through all the messages
func (s State) updateItem(msg actor.Message){
	// update state
}
```

## Send messages between apps 
Different apps can be connected together exchanging messages via [NATS](https://github.com/nats-io/nats.go) in the same way they use locally within the app. You need to configure a NATS server connection, create a registry of exchanged message body types, and give a name to the app for matching with the outboundArea property of a message when initialize Postman.

That's the code in app A
```go
natsToken := os.Getenv("NATS_SECRET")
nc, _:= nats.Connect(nats.DefaultURL, nats.Token(natsToken))
typesRegistry := actor.EnvelopePayloadTypeRegistry{
	"message.WelcomeBody": reflect.TypeOf(message.WelcomeBody{}),
}
actor.InitPostman(actor.WithOutboundMessageService("app-A", nc, reg))
```


That's the code in app B
```go
fromAddress := actor.NewAddress("local", "actor-two")
toAddress := actor.NewOutboundAddress("app-a", "local", "actor-one")
remoteMsg := actor.NewMessage(
	toAddress,
	fromAddress,
	message.WelcomeBody{Text: "hello app A, I'm app B"},
)

err = actor.SendMessage(remoteMsg)
```
