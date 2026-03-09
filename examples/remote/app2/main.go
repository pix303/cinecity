package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pix303/cinecity/pkg/actor"
)

type MainActorTwo struct {
	lastMsg actor.Message
}

var MainActorTwoAddress = actor.NewAddress("local", "actor-two")

func NewMainActorTwo() *MainActorTwo {
	return &MainActorTwo{
		lastMsg: actor.EmptyMessage,
	}
}

func (a *MainActorTwo) Process(msg actor.Message) {
	a.lastMsg = msg
	slog.Info("app two", slog.String("lastmsg", a.lastMsg.String()))
}

func (a *MainActorTwo) GetState() actor.Message {
	return a.lastMsg
}

func (a *MainActorTwo) Shutdown() {
	slog.Info("shutdown app two")
}

type MsgBody struct {
	Text string `json:"text"`
}

func main() {
	natsToken := os.Getenv("NATS_SECRET")
	nc, err := nats.Connect(nats.DefaultURL, nats.Token(natsToken))
	if err != nil {
		panic(err.Error())
	}

	reg := actor.EnvelopePayloadTypeRegistry{}
	actor.InitPostman(actor.WithOutboundMessageService("app2", nc, reg))

	slog.Info("app two")

	fromAddress := actor.NewAddress("local", "actor-two")
	toAddress := actor.NewOutboundAddress("app1", "local", "actor-one")
	rmsg := actor.NewMessage(
		toAddress,
		fromAddress,
		MsgBody{Text: "hello from app two"},
	)

	err = actor.SendMessage(rmsg)
	if err != nil {
		panic(err.Error())
	}

	<-time.After(30 * time.Second)

	slog.Info("end --- app two")
}
