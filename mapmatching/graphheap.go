package mapmatching

import "EVdata/common"

type HeapItem struct {
	node     *common.GraphNode
	distance float64
	prevNode *common.GraphNode
}

func NewHeapItem(n *common.GraphNode, distance float64, prev *common.GraphNode) *HeapItem {
	return &HeapItem{
		node:     n,
		distance: distance,
		prevNode: prev,
	}
}

type GraphHeap struct {
	items []*HeapItem
}

func NewGraphHeap() *GraphHeap {
	return &GraphHeap{
		items: make([]*HeapItem, 0),
	}
}

func (h *GraphHeap) Len() int {
	return len(h.items)
}

func (h *GraphHeap) Push(item *HeapItem) {
	h.items = append(h.items, item)
	h.up(len(h.items) - 1)
}

func (h *GraphHeap) Pop() *HeapItem {
	if len(h.items) == 0 {
		return nil
	}

	item := h.items[0]

	h.items[0], h.items[len(h.items)-1] = h.items[len(h.items)-1], h.items[0]
	h.items = h.items[:len(h.items)-1]

	if len(h.items) > 0 {
		h.down(0)
	}

	return item
}

func (h *GraphHeap) up(index int) {
	if index <= 0 {
		return
	}
	parentIdx := (index - 1) / 2
	if h.items[parentIdx].distance <= h.items[index].distance {
		return
	}
	h.items[parentIdx], h.items[index] = h.items[index], h.items[parentIdx]
	h.up(parentIdx)
}

func (h *GraphHeap) down(index int) {
	smallest := index
	leftIdx := 2*index + 1
	rightIdx := 2*index + 2

	if leftIdx < len(h.items) && h.items[leftIdx].distance < h.items[smallest].distance {
		smallest = leftIdx
	}
	if rightIdx < len(h.items) && h.items[rightIdx].distance < h.items[smallest].distance {
		smallest = rightIdx
	}

	if smallest == index {
		return
	}

	h.items[index], h.items[smallest] = h.items[smallest], h.items[index]
	h.down(smallest)
}
