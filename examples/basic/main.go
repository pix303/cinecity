package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/pix303/cinecity/pkg/actor"
)

// Product represents a product in the inventory.
type Product struct {
	Code     string
	Quantity int
}

// ProductsState represents the state of the products inventory.
type ProductsState struct {
	products []Product
}

func NewProductState() *ProductsState {
	initState := ProductsState{
		products: make([]Product, 0),
	}
	return &initState
}

func (state *ProductsState) getProduct(code string) *Product {
	for i := range state.products {
		cp := &state.products[i]
		if cp.Code == code {
			return cp
		}
	}
	return nil
}

// Messages payload to process

type AddNewProductPayload struct {
	Product Product
}

type AddQuantityToProductPayload struct {
	Product Product
}

type RemoveQuantityToProductPayload struct {
	Product Product
}

type GetProductPayload struct {
	ProductID string
}

type GetProductResponsePayload struct {
	Product Product
}

// Process processes incoming messages and updates the state accordingly.
func (state *ProductsState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {

	case AddNewProductPayload:
		slog.Info("handling add product message")
		p := payload.Product
		state.products = append(state.products, p)

	case AddQuantityToProductPayload:
		slog.Info("handling add quantity to product message")
		p := payload.Product
		cp := state.getProduct(p.Code)
		if cp != nil {
			cp.Quantity += p.Quantity
			slog.Info("update quantity with success", slog.Int("qty", p.Quantity))
		} else {
			slog.Info("update quantity fail")
		}

	case RemoveQuantityToProductPayload:
		slog.Info("handling remove quantity to product message")
		p := payload.Product
		cp := state.getProduct(p.Code)
		if cp != nil {
			cp.Quantity -= p.Quantity
			slog.Info("update quantity with success", slog.Int("qty", p.Quantity))
		} else {
			slog.Info("update quantity fail")
		}

	case GetProductPayload:
		slog.Info("handling retrive product and respond")
		pid := payload.ProductID
		cp := state.getProduct(pid)
		if msg.WithResponse {
			if cp != nil {
				responseMsg := actor.NewReturnMessage(GetProductResponsePayload{Product: *cp}, msg, nil)
				msg.ResponseChan <- responseMsg
			} else {
				responseErrorMsg := actor.NewReturnMessage(nil, msg, errors.New("product not found"))
				msg.ResponseChan <- responseErrorMsg
			}
		}

	default:
		slog.Warn("currently this msg is not handled", slog.String("msg", msg.String()))
	}

	slog.Info("---------------------------------------------------")
	slog.Info("-- Quantity of items in first product", slog.Int("num", state.products[0].Quantity))
	slog.Info("---------------------------------------------------")
}

func (state *ProductsState) GetState() any {
	return state.products
}

// Shutdown cleans up the state when the actor is shutting down.
func (state *ProductsState) Shutdown() {
	state.products = make([]Product, 0)
	slog.Info("all product cleaned")
}

func main() {
	slog.Info("---- start of basic example -----")

	warehouseAddress := actor.NewAddress("local", "warehouse")

	warehouseActor, err := actor.RegisterActor(
		warehouseAddress,
		NewProductState(),
	)

	if err != nil {
		slog.Error("error on create actor", slog.String("err", err.Error()))
		os.Exit(1)
	}

	msg := actor.NewMessage(
		warehouseAddress,
		nil,
		AddNewProductPayload{Product{Code: "ABC", Quantity: 5}},
	)
	msg2 := actor.NewMessage(
		warehouseAddress,
		nil,
		AddQuantityToProductPayload{Product{Code: "ABC", Quantity: 10}},
	)
	msg3 := actor.NewMessage(
		warehouseAddress,
		nil,
		RemoveQuantityToProductPayload{Product{Code: "ABC", Quantity: 2}},
	)

	actor.SendMessage(msg)
	actor.SendMessage(msg2)
	actor.SendMessage(msg3)

	<-time.After(1 * time.Second)

	slog.Info("final state", slog.Any("products", warehouseActor.GetState()))

	msg4 := actor.NewMessageWithResponse(
		warehouseAddress,
		nil,
		GetProductPayload{ProductID: "ABC"},
	)

	response, err := actor.SendMessageWithResponse[GetProductResponsePayload](msg4)
	if err != nil {
		slog.Error("error on get product", slog.String("err", err.Error()))
	} else {
		slog.Info("retrived product", slog.Any("product", response.Product))
	}

	slog.Info("---- end of basic example -------")
}
