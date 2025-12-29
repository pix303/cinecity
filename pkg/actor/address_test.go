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
	assert.Contains(t, a1.String(), "local/a1")
	assert.Contains(t, a2.String(), "local/a2")
}
