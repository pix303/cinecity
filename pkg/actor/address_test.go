package actor_test

import (
	"testing"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/stretchr/testify/assert"
)

func TestPID(t *testing.T) {
	a1 := actor.NewAddress("local", "a1")
	a1bis := actor.NewAddress("local", "a1")
	a2 := actor.NewAddress("local", "a2")
	assert.True(t, a1.IsEqual(a1bis))
	assert.False(t, a2.IsEqual(a1))
	assert.Contains(t, a1.String(), "local.a1")
	assert.Contains(t, a2.String(), "local.a2")
}

func TestNewOutboundAddress(t *testing.T) {
	addr := actor.NewOutboundAddress("remote", "area1", "id1")
	assert.Equal(t, "area1", addr.Area())
	assert.Equal(t, "id1", addr.ID())
	assert.True(t, addr.IsOutbound())
	assert.False(t, addr.IsInbound())
}

func TestAddressArea(t *testing.T) {
	addr := actor.NewAddress("testarea", "testid")
	assert.Equal(t, "testarea", addr.Area())
}

func TestAddressID(t *testing.T) {
	addr := actor.NewAddress("testarea", "testid")
	assert.Equal(t, "testid", addr.ID())
}

func TestAddressIsInbound(t *testing.T) {
	localAddr := actor.NewAddress("local", "test")
	remoteAddr := actor.NewOutboundAddress("remote", "area", "id")
	assert.True(t, localAddr.IsInbound())
	assert.False(t, remoteAddr.IsInbound())
}

func TestGetOutboundAreaPrefix(t *testing.T) {
	prefix := actor.GetOutboundAreaPrefix("myremote")
	assert.Equal(t, "cinecity.myremote", prefix)
}

func TestAddressStringWithOutbound(t *testing.T) {
	addr := actor.NewOutboundAddress("remote", "area1", "id1")
	str := addr.String()
	assert.Equal(t, "cinecity.remote.area1.id1", str)
}

func TestAddressStringNil(t *testing.T) {
	var addr *actor.Address
	assert.Equal(t, "address nil", addr.String())
}

func TestAddressIsSameArea(t *testing.T) {
	addr := actor.NewAddress("testarea", "testid")
	sameArea := "testarea"
	diffArea := "otherarea"
	assert.True(t, addr.IsSameArea(&sameArea))
	assert.False(t, addr.IsSameArea(&diffArea))
}
