package main

import (
	"log/slog"
	"os"
	"reflect"

	"time"

	"github.com/nats-io/nats.go"
	"github.com/pix303/cinecity/pkg/actor"
)

type MainActorOne struct {
	lastMsg actor.Message
}

var MainActorOneAddress = actor.NewAddress("local", "actor-one")

func NewMainActorOne() *MainActorOne {
	return &MainActorOne{
		lastMsg: actor.EmptyMessage,
	}
}

func (a *MainActorOne) Process(msg actor.Message) {
	a.lastMsg = msg
	slog.Info("app one", slog.String("lastmsg", a.lastMsg.String()))
}

func (a *MainActorOne) GetState() any {
	return a.lastMsg
}

func (a *MainActorOne) Shutdown() {
	slog.Info("shutdown main actor one state")
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

	reg := actor.EnvelopePayloadTypeRegistry{
		"main.MsgBody": reflect.TypeOf(MsgBody{}),
	}
	actor.InitPostman(actor.WithOutboundMessageService("app1", nc, reg))

	ma := NewMainActorOne()
	_, err = actor.RegisterActor(MainActorOneAddress, ma)
	if err != nil {
		panic(err.Error())
	}

	slog.Info("app one")
	<-time.After(30 * time.Second)

	slog.Info("end --- app one")
}
