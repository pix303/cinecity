package actor

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/nats-io/nats.go"
)

var (
	ErrAddressInvalid                  = errors.New("address is invalid: area or id are empty")
	ErrActorNotFound                   = errors.New("actor not found")
	ErrActorAddressAlreadyRegistered   = errors.New("actor address already registered")
	ErrInboxReturnMessageBodyTypeWrong = errors.New("return body message type is wrong")
)

type Postman struct {
	actors                 map[string]*Actor
	context                context.Context
	cancelFunc             func()
	enableOutboundMessages bool
	outboundConnection     *nats.Conn
}
type PostmanOption func(*Postman)

func WithOutboundMessage(natsConnection *nats.Conn) PostmanOption {
	return func(p *Postman) {
		p.outboundConnection = natsConnection
		p.enableOutboundMessages = true
	}
}

func OutboundMessageHandler(msg *nats.Msg) {
	slog.Info("outbound message received", slog.String("msg", string(msg.Data)))
	rawAddressSource := strings.Split(msg.Subject, "/")
	if len(rawAddressSource) != 3 {
		slog.Error("outbound message subject is invalid", slog.String("msg", string(msg.Subject)))
		return
	}
	address := NewAddress(
		rawAddressSource[1],
		rawAddressSource[2],
	)

	var body any

	err := json.Unmarshal(msg.Data, body)
	if err != nil {
		slog.Error("outbound message body is invalid", slog.String("err", err.Error()))
		return
	}

	finalMsg := NewMessage(
		address,
		nil,
		body,
	)

	err = SendMessage(finalMsg)
	if err != nil {
		slog.Error("outbound message fail to be send", slog.String("err", err.Error()))
	}
}

var instance *Postman
var onceGuard sync.Once

// InitPostman initialize postman service (singleton)
func InitPostman(opts ...PostmanOption) *Postman {
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

		instance = &Postman{
			actors:     make(map[string]*Actor, 10),
			context:    ctx,
			cancelFunc: cancFunc,
		}

		natsToken := os.Getenv("NATS_SECRET")
		if natsToken == "" {
			slog.Warn("NATS_SECRET is not set, so outbound feature is not active")
			return
		} else {
			nc, err := nats.Connect(nats.DefaultURL, nats.Token(natsToken))
			if err != nil {
				slog.Warn("NATS service fail on init", slog.String("err", err.Error()))
			}

			// TODO: check how to use subscription
			_, err = nc.Subscribe(OutboundPrefix, OutboundMessageHandler)
			if err != nil {
				slog.Warn("NATS service fail on init", slog.String("err", err.Error()))
			}
			slog.Info("NATS service is active")
		}

		for _, opt := range opts {
			opt(instance)
		}
	})

	return instance
}

func GetPostman() *Postman {
	if instance == nil {
		panic("postman must be initialized before")
	}
	return instance
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

	if msg.To.isOutbound {
		var jsonResource []byte
		err := json.Unmarshal(jsonResource, msg.Body)
		if err != nil {
			slog.Error("outbound message body must be json", slog.String("msg", msg.String()))
			return err
		}
		err = p.outboundConnection.Publish(OutboundPrefix, jsonResource)
		if err != nil {
			slog.Error("outbound error on publish", slog.String("err", err.Error()))
			return err
		}
		return err
	}

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
	p.outboundConnection.Close()
	p.cancelFunc()
}

func NumActors() int {
	p := GetPostman()
	return len(p.actors)
}
