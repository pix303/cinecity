package actor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/stretchr/testify/assert"
)

type TestPayload struct {
	Code string `json:"code"`
	Pcs  int    `json:"pcs"`
}

func TestNewOutboundEnvelope(t *testing.T) {
	body := TestPayload{Code: "test", Pcs: 25}
	bodyType := "test.payload"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.NoError(t, err)
	assert.Equal(t, bodyType, envelope.BodyType)
	assert.NotEmpty(t, envelope.RawBody)

	var unmarshaled TestPayload
	err = json.Unmarshal(envelope.RawBody, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, body, unmarshaled)
}

func TestNewOutboundEnvelopeWithStringBody(t *testing.T) {
	body := "simple string body"
	bodyType := "string"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.NoError(t, err)
	assert.Equal(t, bodyType, envelope.BodyType)
	assert.Equal(t, `"simple string body"`, string(envelope.RawBody))
}

func TestNewOutboundEnvelopeWithMapBody(t *testing.T) {
	body := map[string]any{"key": "value", "number": 42}
	bodyType := "map"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.NoError(t, err)
	assert.Equal(t, bodyType, envelope.BodyType)

	var unmarshaled map[string]any
	err = json.Unmarshal(envelope.RawBody, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "value", unmarshaled["key"])
	assert.Equal(t, float64(42), unmarshaled["number"])
}

func TestNewOutboundEnvelopeWithStructPointer(t *testing.T) {
	body := &TestPayload{Code: "pointer", Pcs: 30}
	bodyType := "pointer.payload"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.NoError(t, err)
	assert.Equal(t, bodyType, envelope.BodyType)

	var unmarshaled TestPayload
	err = json.Unmarshal(envelope.RawBody, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "pointer", unmarshaled.Code)
	assert.Equal(t, 30, unmarshaled.Pcs)
}

func TestNewOutboundEnvelopeEmptyBody(t *testing.T) {
	body := struct{}{}
	bodyType := "empty"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.NoError(t, err)
	assert.Equal(t, bodyType, envelope.BodyType)
	assert.Equal(t, "{}", string(envelope.RawBody))
}

func TestNewOutboundEnvelopeNilBody(t *testing.T) {
	var body interface{} = nil
	bodyType := "nil"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.NoError(t, err)
	assert.Equal(t, bodyType, envelope.BodyType)
	assert.Equal(t, "null", string(envelope.RawBody))
}

func TestOutboundEvenlopeFields(t *testing.T) {
	body := "test"
	bodyType := "type1"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.NoError(t, err)
	assert.NotZero(t, envelope.BodyType)
	assert.NotZero(t, envelope.RawBody)
}

type FailingMarshal struct {
	Value string
}

var errCustom = errors.New("custom marshal error")

func (f FailingMarshal) MarshalJSON() ([]byte, error) {
	return nil, errCustom
}

func TestNewOutboundEnvelopeMarshalError(t *testing.T) {
	body := FailingMarshal{Value: "fail"}
	bodyType := "fail.payload"

	envelope, err := actor.NewOutboundEnvelope(body, bodyType)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json")
	assert.Equal(t, "", envelope.BodyType)
	assert.Empty(t, envelope.RawBody)
}
