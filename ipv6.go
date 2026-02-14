package addrtrie

import (
	"encoding/binary"
	"errors"
	"net"
)

type ipv6Addr struct {
	hi uint64 // high 64 bits
	lo uint64 // low 64 bits
}

func newIPv6Addr(ip net.IP) ipv6Addr {
	ip16 := ip.To16()
	hi := binary.BigEndian.Uint64(ip16[0:8])
	lo := binary.BigEndian.Uint64(ip16[8:16])
	return ipv6Addr{hi: hi, lo: lo}
}

func (a ipv6Addr) getBit(i int) int {
	if i < 64 {
		shift := 63 - i
		return int((a.hi >> shift) & 1)
	}
	shift := 63 - (i - 64)
	return int((a.lo >> shift) & 1)
}

func parseIPorCIDRIPv6(s string) (ipv6Addr, int, error) {
	if ip, ipNet, err := net.ParseCIDR(s); err == nil {
		ip = ip.To16()
		if ip == nil || ip.To4() != nil {
			return ipv6Addr{}, 0, net.InvalidAddrError("non-IPv6 CIDR")
		}
		ones, bits := ipNet.Mask.Size()
		if bits != 128 {
			return ipv6Addr{}, 0, net.InvalidAddrError("non-IPv6 mask")
		}
		if ones == 0 && bits == 0 {
			return ipv6Addr{}, 0, net.InvalidAddrError("non-canonical mask")
		}
		return newIPv6Addr(ip), ones, nil
	}

	ip := net.ParseIP(s).To16()
	if ip == nil || ip.To4() != nil {
		return ipv6Addr{}, 0, net.InvalidAddrError("invalid IPv6 address")
	}
	return newIPv6Addr(ip), 128, nil
}

type bitNode6[T any] struct {
	children    [2]*bitNode6[T]
	value       T
	valueExists bool
}

type BitTrie6[T any] struct {
	root *bitNode6[T]
}

func NewBitTrie6[T any]() *BitTrie6[T] {
	return &BitTrie6[T]{root: &bitNode6[T]{}}
}

func (t *BitTrie6[T]) Insert(prefix string, value T) error {
	if prefix == "*" {
		t.root.value = value
		t.root.valueExists = true
		return nil
	}
	addr, bitLen, err := parseIPorCIDRIPv6(prefix)
	if err != nil {
		return err
	}
	if bitLen < 0 || bitLen > 128 {
		return errors.New("invalid prefix length")
	}

	cur := t.root
	for i := range bitLen {
		b := addr.getBit(i)
		if cur.children[b] == nil {
			cur.children[b] = &bitNode6[T]{}
		}
		cur = cur.children[b]
	}
	cur.value = value
	cur.valueExists = true
	return nil
}

func (t *BitTrie6[T]) Find(ipStr string) (matched T, exists bool) {
	ip := net.ParseIP(ipStr).To16()
	if ip == nil || ip.To4() != nil {
		return
	}
	addr := newIPv6Addr(ip)

	cur := t.root
	for i := range 128 {
		if cur.valueExists {
			matched = cur.value
			exists = true
		}
		b := addr.getBit(i)
		if cur.children[b] == nil {
			break
		}
		cur = cur.children[b]
	}
	if cur.valueExists {
		matched = cur.value
		exists = true
	}
	if !exists && t.root.valueExists {
		return t.root.value, true
	}
	return
}
