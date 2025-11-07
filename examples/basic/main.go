package main

import (
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

// Process processes incoming messages and updates the state accordingly.
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
		if cp != nil {
			cp.Quantity += p.Quantity
			slog.Info("AddQuantityProductMsg update with success", slog.Int("qty", p.Quantity))
		} else {
			slog.Info("AddQuantityProductMsg update fail")
		}

	case RemoveQuantityToProductPayload:
		slog.Info("RemoveQuantityProductMsg")
		p := payload.Product
		cp := state.getProduct(p.Code)
		if cp != nil {
			cp.Quantity -= p.Quantity
			slog.Info("RemoveQuantityProductMsg update with success", slog.Int("qty", p.Quantity))
		} else {
			slog.Info("RemoveQuantityProductMsg update fail")
		}

	default:
		slog.Warn("this msg is unknown", slog.String("msg", msg.String()))
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
	slog.Info("---- end of basic example -------")
}
