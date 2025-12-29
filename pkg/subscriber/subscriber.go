package subscriber

import (
	"log/slog"

	"github.com/pix303/cinecity/pkg/actor"
)

type SubscriptionsActor struct {
	subscribers []*actor.Address
}

func NewSubscriptionsActor() *SubscriptionsActor {
	return &SubscriptionsActor{
		subscribers: make([]*actor.Address, 0),
	}
}

type AddSubscriptionMessageBody struct{}

func NewAddSubcriptionMessage(subscriberAddress *actor.Address, notifierAddress *actor.Address) actor.Message {
	return actor.Message{
		From: subscriberAddress,
		To:   notifierAddress,
		Body: AddSubscriptionMessageBody{},
	}
}

type RemoveSubscriptionMessageBody struct{}

func NewRemoveSubscriptionMessage(subscriberAddress *actor.Address, notifierAddress *actor.Address) actor.Message {
	return actor.Message{
		From: subscriberAddress,
		To:   notifierAddress,
		Body: RemoveSubscriptionMessageBody{},
	}
}

func (actor *SubscriptionsActor) Process(msg actor.Message) {
	switch msg.Body.(type) {
	case AddSubscriptionMessageBody:
		actor.addSubscription(msg.From)
	case RemoveSubscriptionMessageBody:
		actor.removeSubscription(msg.From)
	}
}

func NewSubscribersMessage(from *actor.Address, body any) actor.Message {
	return actor.Message{
		From: from,
		Body: body,
	}
}

func (state *SubscriptionsActor) addSubscription(subscriberAddress *actor.Address) {
	state.subscribers = append(state.subscribers, subscriberAddress)
}

func (state *SubscriptionsActor) removeSubscription(subscriberAddress *actor.Address) {
	for i, v := range state.subscribers {
		if v.IsEqual(subscriberAddress) {
			state.subscribers = append(state.subscribers[:i], state.subscribers[i+1:]...)
		}
	}
}

func (state *SubscriptionsActor) NumSubscribers() int {
	return len(state.subscribers)
}

func (state *SubscriptionsActor) NotifySubscribers(msg actor.Message) int {
	result := 0
	for _, sub := range state.subscribers {
		msg.To = sub
		slog.Info("sending msg to subscriber", slog.String("msg", msg.String()), slog.String("subscriber", sub.String()))
		err := actor.SendMessage(msg)
		if err != nil {
			result++
			slog.Warn("error on send msg to subscribers", slog.String("msg", msg.String()), slog.String("err", err.Error()))
		}
	}
	return result
}
