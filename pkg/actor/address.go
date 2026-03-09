package actor

import (
	"fmt"
	"strings"
)

type Address struct {
	area         string
	id           string
	outboundArea string
}

// NewAddress creates a new address with the given area and id for sending message within the application (see outbound address to send message to remote application).
func NewAddress(area, id string) *Address {
	return &Address{
		area,
		id,
		"",
	}
}

const OutboundPrefix string = "cinecity"
const AddressSeparator string = "."

// NewOutboundAddress creates a new address with the given area and id for sending message to remote application.
func NewOutboundAddress(outboundArea, area, id string) *Address {
	return &Address{
		area,
		id,
		outboundArea,
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
	addrParts := []string{
		addr.area,
		addr.id,
	}

	if addr.outboundArea != "" {
		prefix := []string{GetOutboundAreaPrefix(addr.outboundArea)}
		addrParts = append(prefix, addrParts...)
	}
	return strings.Join(addrParts, AddressSeparator)
}

func (addr *Address) Area() string {
	return addr.area
}

func (addr *Address) ID() string {
	return addr.id
}

func (addr *Address) IsOutbound() bool {
	return addr.outboundArea != ""
}

func (addr *Address) IsInbound() bool {
	return !addr.IsOutbound()
}

func GetOutboundAreaPrefix(outboundArea string) string {
	return fmt.Sprintf("%s.%s", OutboundPrefix, outboundArea)
}
