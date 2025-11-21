package actor

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	ErrAddressInvalid                  = errors.New("address is invalid: area or id are empty")
	ErrActorNotFound                   = errors.New("actor not found")
	ErrActorAddressAlreadyRegistered   = errors.New("actor address already registered")
	ErrInboxReturnMessageBodyTypeWrong = errors.New("return body message type is wrong")
)

type Postman struct {
	actors     map[string]*Actor
	context    context.Context
	cancelFunc func()
}

var instance Postman
var onceGuard sync.Once

func GetPostman() *Postman {
	onceGuard.Do(func() {
		ctx, cancFunc := context.WithCancel(context.Background())
		extCancel := make(chan os.Signal, 1)
		signal.Notify(extCancel, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			for {
				s := <-extCancel
				switch s {
				case syscall.SIGINT, syscall.SIGTERM:
					ShutdownAll()
				}
			}
		}()

		instance = Postman{
			actors:     make(map[string]*Actor, 10),
			context:    ctx,
			cancelFunc: cancFunc,
		}
	})
	return &instance
}

func (postman *Postman) GetContext() context.Context {
	return postman.context
}

func RegisterActor(address *Address, processor StateProcessor) (*Actor, error) {
	if address == nil || address.area == "" || address.id == "" {
		return nil, ErrAddressInvalid
	}

	a := Actor{
		address:    address,
		state:      processor,
		MessageBox: make(chan Message, 100),
		isClosed:   true,
	}

	p := GetPostman()
	if temp := p.actors[a.GetAddress().String()]; temp != nil {
		slog.Error(ErrActorAddressAlreadyRegistered.Error(), slog.String("actor-address", a.GetAddress().String()))
		return nil, ErrActorAddressAlreadyRegistered
	}

	p.actors[a.GetAddress().String()] = &a
	slog.Info("actor registered", slog.String("a", a.GetAddress().String()))
	a.Activate()
	return &a, nil
}

func UnRegisterActor(address *Address) {
	p := GetPostman()
	delete(p.actors, address.String())
}

func SendMessage(msg Message) error {
	p := GetPostman()
	actor := p.actors[msg.To.String()]

	if actor == nil {
		slog.Error("actor not found", slog.String("actor-address", msg.To.String()))
		return ErrActorNotFound
	}

	slog.Debug("actor found, sending msg", slog.String("actor-address", msg.To.String()))
	err := actor.Inbox(msg)
	if err != nil {
		slog.Error("actor inbox return error", slog.String("actor-address", msg.To.String()), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func SendMessageWithResponse[T any](msg Message) (T, error) {
	p := GetPostman()
	actor := p.actors[msg.To.String()]
	if actor == nil {
		slog.Error("actor not found", slog.String("actor-address", msg.To.String()))
		return *new(T), ErrActorNotFound
	}

	returnMsg, err := actor.InboxAndWaitResponse(msg)
	if err != nil {
		slog.Error(
			"actor inbox with response return error",
			slog.String("actor-address", msg.To.String()),
			slog.String("error", err.Error()),
			slog.Any("msg", new(T)),
		)
		return *new(T), err
	}

	if body, ok := returnMsg.Body.(T); ok {
		return body, nil
	}

	return *new(T), ErrInboxReturnMessageBodyTypeWrong
}

func BroadcastMessage(msg Message, area *string) int {
	counter := 0
	p := GetPostman()

	for _, a := range p.actors {
		if a.GetAddress().IsEqual(msg.From) {
			continue
		}

		if area != nil && !a.GetAddress().IsSameArea(area) {
			continue
		}

		counter++
		err := a.Inbox(msg)
		if err != nil {
			slog.Warn("actor inbox error on broadcasting message", slog.String("actor-address", msg.To.String()), slog.String("error", err.Error()))
		}
	}

	return counter
}

func ShutdownAll() {
	p := GetPostman()
	for _, a := range p.actors {
		a.Drop()
	}
	p.actors = make(map[string]*Actor)
	p.cancelFunc()
}

func NumActors() int {
	p := GetPostman()
	return len(p.actors)
}
