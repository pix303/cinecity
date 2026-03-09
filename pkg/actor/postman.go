package actor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
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
	ErrOutboundMessageBodyMustBeNotNil = errors.New("outbound message body must be not nil")
)

type Postman struct {
	actors                 map[string]*Actor
	context                context.Context
	cancelFunc             func()
	enableOutboundMessages bool
	outboundOptions        *OutboundOptions
}

type PostmanOption func(*Postman)

func WithOutboundMessageService(
	outboundArea string,
	natsConnection *nats.Conn,
	payloadTypeRegistry EnvelopePayloadTypeRegistry,
) PostmanOption {
	return func(p *Postman) {
		p.enableOutboundMessages = true

		oo := OutboundOptions{
			outboundArea:   outboundArea,
			natsConnection: natsConnection,
			typeRegistry:   payloadTypeRegistry,
		}
		p.outboundOptions = &oo
	}
}

func (p *Postman) OutboundMessageHandler(msg *nats.Msg) {
	slog.Info("outbound message received", slog.String("msg", string(msg.Data)), slog.String("subj", string(msg.Subject)))
	rawAddressSource := strings.Split(msg.Subject, AddressSeparator)
	if len(rawAddressSource) < 3 {
		slog.Error("outbound message subject is invalid", slog.Any("parts", rawAddressSource))
		return
	}

	localActorAddress := NewAddress(
		rawAddressSource[2],
		rawAddressSource[3],
	)

	var envelop OutboundEvenlope

	err := json.Unmarshal(msg.Data, &envelop)
	if err != nil {
		slog.Error("outbound message payload is invalid", slog.String("err", err.Error()))
		return
	}

	payloadType := p.outboundOptions.typeRegistry[envelop.BodyType]
	if payloadType == nil {
		slog.Error("outbound payload type not found in registry", slog.String("type", envelop.BodyType))
		return
	}

	payload := reflect.New(payloadType).Interface()
	err = json.Unmarshal(envelop.RawBody, payload)
	if err != nil {
		slog.Error("outbound message payload is invalid", slog.String("err", err.Error()))
		return
	}

	slog.Info("outbound message envelop", slog.Any("payload", payload))

	finalMsg := NewMessage(
		localActorAddress,
		nil,
		envelop,
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

		for _, opt := range opts {
			opt(instance)
		}

		if instance.enableOutboundMessages {
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
				subj := fmt.Sprintf("%s.*.*", GetOutboundAreaPrefix(instance.outboundOptions.outboundArea))
				_, err = nc.Subscribe(subj, instance.OutboundMessageHandler)
				if err != nil {
					slog.Warn("NATS service fail on init", slog.String("err", err.Error()))
				}
				slog.Info("NATS service is active")
			}
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

	if msg.To.IsOutbound() {
		if msg.Body == nil {
			return ErrOutboundMessageBodyMustBeNotNil
		}

		bodyType := reflect.TypeOf(msg.Body).String()
		envelop, err := NewOutboundEnvelope(msg.Body, bodyType)
		if err != nil {
			return err
		}

		envelopPayload, err := json.Marshal(envelop)
		if err != nil {
			return err
		}

		slog.Info("outbound message", slog.String("to", msg.To.String()))
		err = p.outboundOptions.natsConnection.Publish(msg.To.String(), envelopPayload)
		if err != nil {
			slog.Error("outbound error on publish", slog.String("err", err.Error()))
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

		err := a.Inbox(msg)
		if err != nil {
			slog.Warn("actor inbox error on broadcasting message", slog.String("actor-address", msg.To.String()), slog.String("error", err.Error()))
			continue
		}
		counter++
	}

	return counter
}

func ShutdownAll() {
	p := GetPostman()
	for _, a := range p.actors {
		a.Drop()
	}
	p.actors = make(map[string]*Actor)

	if p.enableOutboundMessages && p.outboundOptions != nil && p.outboundOptions.natsConnection != nil {
		p.outboundOptions.natsConnection.Close()
	}

	p.cancelFunc()
}

func NumActors() int {
	p := GetPostman()
	return len(p.actors)
}
