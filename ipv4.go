package addrtrie

import (
	"net"
	"encoding/binary"
)

func parseIPorCIDR(s string) (ip uint32, bitLen int, err error) {
	if _, ipNet, e := net.ParseCIDR(s); e == nil {
		ip = binary.BigEndian.Uint32(ipNet.IP.To4())
		ones, bits := ipNet.Mask.Size()
		if bits != 32 {
			return 0, 0, net.InvalidAddrError("non-IPv4 mask")
		}
		if ones == 0 && bits == 0 {
			return 0, 0, net.InvalidAddrError("non-canonical mask")
		}
		return ip, ones, nil
	}
	parsed := net.ParseIP(s).To4()
	if parsed == nil {
		return 0, 0, net.InvalidAddrError("invalid IPv4 address")
	}
	ip = binary.BigEndian.Uint32(parsed)
	return ip, 32, nil
}

func getBit(v uint32, i int) int {
	shift := 31 - i
	return int((v >> shift) & 1)
}

type bitNode[T any] struct {
	children [2]*bitNode[T]
	value    T
	valueExists bool
}

type BitTrie[T any] struct {
	root *bitNode[T]
}

func NewBitTrie[T any]() *BitTrie[T] {
	return &BitTrie[T]{root: &bitNode[T]{children: [2]*bitNode[T]{}}}
}

func (t *BitTrie[T]) Insert(prefix string, value T) error {
	if prefix == "*" {
		t.root.value = value
		t.root.valueExists = true
	}
	ip, bitLen, err := parseIPorCIDR(prefix)
	if err != nil {
		return err
	}

	cur := t.root
	for i := range bitLen {
		b := getBit(ip, i)
		if cur.children[b] == nil {
			cur.children[b] = &bitNode[T]{children: [2]*bitNode[T]{}}
		}
		cur = cur.children[b]
	}
	cur.value = value
	return nil
}

func (t *BitTrie[T]) Find(ipStr string) (matched T, exists bool) {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return
	}
	ipUint := binary.BigEndian.Uint32(ip)

	cur := t.root
	for i := range 32 {
		if cur.valueExists {
			matched = cur.value
			exists = true
		}
		b := getBit(ipUint, i)
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
		return t.root.value, t.root.valueExists
	}
	return
}