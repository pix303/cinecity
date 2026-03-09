package actor

import (
	"encoding/json"
	"log/slog"
	"reflect"

	"github.com/nats-io/nats.go"
)

type OutboundEvenlope struct {
	BodyType string `json:"bodyType"`
	RawBody  []byte `json:"rawBody"`
}

func NewOutboundEnvelope(body any, bodyType string) (OutboundEvenlope, error) {
	rawbody, err := json.Marshal(body)
	if err != nil {
		slog.Error("fail to marshal body", slog.String("err", err.Error()))
		return OutboundEvenlope{}, err
	}

	return OutboundEvenlope{
		BodyType: bodyType,
		RawBody:  rawbody,
	}, nil
}

type EnvelopePayloadTypeRegistry map[string]reflect.Type

type OutboundOptions struct {
	natsConnection *nats.Conn
	typeRegistry   EnvelopePayloadTypeRegistry
	outboundArea   string
}
