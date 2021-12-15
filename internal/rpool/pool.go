package rpool

import (
	"k8s.io/apimachinery/pkg/labels"
)

type Node struct {
	key       uint32
	allocated bool
	labels    *labels.Set
}

func (n *Node) GetLabels() *labels.Set {
	return n.labels
}

func (n *Node) GetAllocated() bool {
	return n.allocated
}

func (n *Node) GetKey() uint32 {
	return n.key
}

type pool struct {
	size               int
	allocationStrategy string
	nodes              []*Node
}

type Pool interface {
	QueryByLabels(labels.Selector) []uint32
	QueryByIndex(int) []uint32
	Allocate(*uint32, map[string]string) (uint32, bool)
	DeAllocate(uint32)
	GetAllocated() []uint32
}

func New(start, end uint32, allocStrategy string) Pool {
	size := int(end) - int(start)
	nodes := make([]*Node, 0, size)

	for i := 0; i <= int(end)-int(start); i++ {
		nodes = append(nodes, &Node{
			key:       start + uint32(i),
			allocated: false,
			labels:    &labels.Set{},
		})
	}
	p := &pool{
		size:               size,
		allocationStrategy: allocStrategy,
		nodes:              nodes,
	}
	return p
}

func (p *pool) QueryByLabels(selector labels.Selector) []uint32 {
	matches := make([]uint32, 0)
	for _, n := range p.nodes {
		if selector.Matches(n.GetLabels()) {
			matches = append(matches, n.key)
		}
	}
	return matches
}

func (p *pool) QueryByIndex(index int) []uint32 {
	matches := make([]uint32, 0)
	if index < p.size {
		if p.nodes[index].allocated {
			matches = append(matches, p.nodes[index].key)
		}
	}
	return matches
}

func (p *pool) Allocate(key *uint32, label map[string]string) (uint32, bool) {
	// TODO index based allocation
	switch p.allocationStrategy {
	case "deterministic":
		// key is used as index here
		if key != nil && int(*key) < p.size {
			n := p.nodes[int(*key)]
			if n.allocated {
				// TODO check if the labels match; right now we overwrite the labels
				n.allocated = true
				// assign labels
				mergedlabel := labels.Merge(labels.Set(label), *n.labels)
				n.labels = &mergedlabel
				return n.key, true
			} else {
				n.allocated = true
				// assign labels
				mergedlabel := labels.Merge(labels.Set(label), *n.labels)
				n.labels = &mergedlabel
				return n.key, true
			}
		}
		return 0, false
	default:
		// allocation strategy = first-available
		// key is used to match the pool key
		if key == nil {
			// allocate based on first free index
			for _, n := range p.nodes {
				if !n.allocated {
					n.allocated = true
					// assign labels
					mergedlabel := labels.Merge(labels.Set(label), *n.labels)
					n.labels = &mergedlabel
					return n.key, true
				}
			}
		} else {
			for _, n := range p.nodes {
				if n.key == *key {
					n.allocated = true
					// assign labels
					mergedlabel := labels.Merge(labels.Set(label), *n.labels)
					n.labels = &mergedlabel
					return n.key, true
				}
			}
		}
	}
	return 0, false
}

func (p *pool) DeAllocate(key uint32) {
	for _, n := range p.nodes {
		if n.key == key {
			n.allocated = false
			n.labels = &labels.Set{}
		}
	}
}

func (p *pool) GetAllocated() []uint32 {
	allocated := make([]uint32, 0)
	for _, n := range p.nodes {
		if n.allocated {
			allocated = append(allocated, n.key)
		}
	}
	return allocated
}
