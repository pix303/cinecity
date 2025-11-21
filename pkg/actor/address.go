package actor

import (
	"fmt"
)

type Address struct {
	area string
	id   string
}

func NewAddress(area, id string) *Address {
	return &Address{
		area,
		id,
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
	return fmt.Sprintf("%s.%s", addr.area, addr.id)
}
