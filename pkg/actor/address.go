package actor

import (
	"fmt"
)

type Address struct {
	area       string
	id         string
	isOutbound bool
}

// NewAddress creates a new address with the given area and id for sending message within the application (see outbound address to send message to remote application).
func NewAddress(area, id string) *Address {
	return &Address{
		area,
		id,
		false,
	}
}

const OutboundPrefix string = "cinecity.outbound"

// NewOutboundAddress creates a new address with the given area and id for sending message to remote application.
func NewOutboundAddress(area, id string) *Address {
	return &Address{
		area,
		id,
		true,
	}
}

func (addr *Address) IsEqual(address *Address) bool {
	return addr.area == address.area && addr.id == address.id
}

func (addr *Address) IsSameArea(area *string) bool {
	return addr.area == *area
}

func (addr *Address) String() string {
	if addr == nil {
		return "address nil"
	}

	if addr.isOutbound {
		return fmt.Sprintf("%s/%s/%s", OutboundPrefix, addr.area, addr.id)
	}
	return fmt.Sprintf("%s/%s", addr.area, addr.id)
}

func (addr *Address) Area() string {
	return addr.area
}

func (addr *Address) ID() string {
	return addr.id
}

func (addr *Address) IsOutbound() bool {
	return addr.isOutbound
}

func (addr *Address) IsInbound() bool {
	return !addr.isOutbound
}
