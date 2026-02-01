package addrtrie

import (
	"errors"
	"strings"
)

type labelNode[T any] struct {
	children map[string]*labelNode[T]
	value    *T
}

type DomainMatcher[T any] struct {
	exactDomains map[string]*T
	root         *labelNode[T]
}

func NewDomainMatcher[T any]() *DomainMatcher[T] {
	return &DomainMatcher[T]{
		exactDomains: make(map[string]*T),
		root:         &labelNode[T]{children: make(map[string]*labelNode[T])},
	}
}

func (m *DomainMatcher[T]) insertTrie(domain string, value *T) {
	node := m.root
	i := len(domain) - 1
	for i >= 0 {
		j := strings.LastIndexByte(domain[:i+1], '.')
		var label string
		if j == -1 {
			label = domain[:i+1]
			i = -1
		} else {
			label = domain[j+1 : i+1]
			i = j - 1
		}

		if node.children[label] == nil {
			node.children[label] = &labelNode[T]{children: make(map[string]*labelNode[T])}
		}
		node = node.children[label]
	}
	node.value = value
}

func (m *DomainMatcher[T]) Add(pattern string, value T) error {
	if !strings.Contains(pattern, ".") {
		return errors.New("invalid pattern: " + pattern)
	}

	switch {
	case strings.HasPrefix(pattern, "*."):
		m.insertTrie(pattern[2:], &value)
	case strings.HasPrefix(pattern, "*"):
		m.exactDomains[pattern[1:]] = &value
		m.insertTrie(pattern[1:], &value)
	default:
		m.exactDomains[pattern] = &value
	}
	return nil
}

func (m *DomainMatcher[T]) Find(domain string) *T {
	if value, ok := m.exactDomains[domain]; ok {
		return value
	}

	node := m.root
	var matched *T
	for i := len(domain) - 1; ; {
		j := strings.LastIndexByte(domain[:i+1], '.')
		if j == -1 {
			break
		}
		child, ok := node.children[domain[j+1:i+1]]
		if !ok {
			break
		}
		i = j - 1
		node = child
		if node.value != nil {
			matched = node.value
		}
	}
	return matched
}
