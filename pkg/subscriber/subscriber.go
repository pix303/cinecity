package subscriber

import (
	"log/slog"

	"github.com/pix303/cinecity/pkg/actor"
)

type SubscriptionsState struct {
	subscribers []*actor.Address
}

func NewSubscriptionsState() *SubscriptionsState {
	return &SubscriptionsState{
		subscribers: make([]*actor.Address, 0),
	}
}

func (state *SubscriptionsState) AddSubscription(subscriberAddress *actor.Address) {
	state.subscribers = append(state.subscribers, subscriberAddress)
}

func (state *SubscriptionsState) RemoveSubscription(subscriberAddress *actor.Address) {
	for i, v := range state.subscribers {
		if v.IsEqual(subscriberAddress) {
			state.subscribers = append(state.subscribers[:i], state.subscribers[i+1:]...)
		}
	}
}

func (state *SubscriptionsState) NumSubscribers() int {
	return len(state.subscribers)
}

func (state *SubscriptionsState) NotifySubscribers(msg actor.Message) {
	for _, sub := range state.subscribers {
		msg.To = sub
		slog.Info("sending msg to subscriber", slog.String("msg", msg.String()), slog.String("subscriber", sub.String()))
		err := actor.SendMessage(msg)
		if err != nil {
			slog.Warn("error on send msg to subscribers", slog.String("msg", msg.String()), slog.String("err", err.Error()))
		}
	}
}
