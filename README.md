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

The core method to implement for a state processor is Process that takes a message as parameter, evaluate the body type and act consequentially 

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

To create and register and actor

```go

warehouseAddress := actor.NewAddress("local", "warehouse")
warehouseActor, err := actor.RegisterActor(
  warehouseAddress,
	// usually create before because of return of errors due to dependences initialization
	NewProductState(), 
)

// check for already registerd address
```

Messages are shipping in a async mode; here an example to create a message and send it

```go
msg := actor.NewMessage(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Product{Code: "ABC", Quantity: 2}}, // body
	)

err := actor.SendMessage(msg)
	
```

Here an example to send a message and wait a response

```go
msg := actor.NewMessageWithResponse(
		warehouseAddress, // to
		customerAddress, // from
		AddQuantityToProductPayload{Product{Code: "ABC", Quantity: 2}}, // body
	)

response, err := actor.SendMessageWithResponse[AddQuantityToProductResultPayload](msg)
	
```
