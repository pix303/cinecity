package main

import (
	"log/slog"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/cinecity/pkg/subscriber"
)

type SentenceState struct {
	sentence []string
	subs     subscriber.Subscriptions
}

func NewSentenceState() *SentenceState {
	s := SentenceState{
		sentence: make([]string, 0),
		subs:     *subscriber.NewSubscription(),
	}

	return &s
}

type AddWordBody struct {
	Word string
}

func (state *SentenceState) Process(msg actor.Message) {
	state.subs.Process(msg)

	switch payload := msg.Body.(type) {
	case AddWordBody:
		state.sentence = append(state.sentence, payload.Word)
		slog.Info("AddWordBody msg")
	default:
		slog.Warn("no payload type processed")
	}
}

func (state *SentenceState) Shutdown() {
	state.sentence = make([]string, 0)
}

func (state *SentenceState) GetState() any {
	return state.sentence
}

type SentenceObserverState struct {
}

func NewSencenceObserverState() *SentenceObserverState {
	return &SentenceObserverState{}
}

type BigSentenceBody struct{}

func (observer *SentenceObserverState) Process(msg actor.Message) {

	switch payload := msg.Body; payload.(type) {
	case BigSentenceBody:
		slog.Info("the sentence is big! %d words!")
	default:
		slog.Warn("no payload type processed")
	}
}

func (state *SentenceObserverState) Shutdown() {
}

func (state *SentenceObserverState) GetState() any {
	return nil
}

func main() {

	sentenceActorAddress := actor.NewAddress("local", "sentence")
	_, err := actor.RegisterActor(
		sentenceActorAddress,
		NewSentenceState(),
	)
	if err != nil {
		panic(err)
	}

	sentenceObserverActorAddress := actor.NewAddress("local", "sentence-observer-1")
	_, err = actor.RegisterActor(
		sentenceObserverActorAddress,
		NewSencenceObserverState(),
	)

	if err != nil {
		panic(err)
	}

	subMsg := subscriber.NewAddSubcriptionMessage(sentenceObserverActorAddress, sentenceActorAddress)
	err = actor.SendMessage(subMsg)

	if err != nil {
		panic(err)
	}

	addWordMsg := actor.NewMessage(sentenceActorAddress, nil, AddWordBody{Word: "hello"})
	err = actor.SendMessage(addWordMsg)

	if err != nil {
		panic(err)
	}

	addWordMsg2 := actor.NewMessage(sentenceActorAddress, nil, AddWordBody{Word: "world"})
	err = actor.SendMessage(addWordMsg2)

	if err != nil {
		panic(err)
	}
}
