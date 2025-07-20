package kway

import (
	"container/heap"
	"iter"
)

type mergeState[T interface{ index() int }] struct {
	cmp   func(a, b T) int
	seqs  []iter.Seq[T]
	items []T
}

func (x *mergeState[T]) Len() int { return len(x.items) }

func (x *mergeState[T]) Less(i, j int) bool {
	if v := x.cmp(x.items[i], x.items[j]); v != 0 {
		return v < 0
	}
	// fall back to comparison by index (documented behavior)
	return x.items[i].index() < x.items[j].index()
}

func (x *mergeState[T]) Swap(i, j int) {
	x.items[i], x.items[j] = x.items[j], x.items[i]
}

func (x *mergeState[T]) Push(v any) {
	x.items = append(x.items, v.(T))
}

func (x *mergeState[T]) Pop() (item any) {
	old := x.items
	i := len(old) - 1
	item = old[i]
	old[i] = *new(T)
	x.items = old[:i]
	return item
}

func (x *mergeState[T]) all(yield func(T) bool) {
	x.items = make([]T, 0, len(x.seqs))
	pulls := make([]func() (T, bool), len(x.seqs))
	for i, seq := range x.seqs {
		if seq != nil {
			next, stop := iter.Pull(seq)
			defer stop()
			if v, ok := next(); ok {
				x.items = append(x.items, v)
				pulls[i] = next
			}
		}
	}
	heap.Init(x)
	for len(x.items) != 0 {
		v := heap.Pop(x).(T)
		if !yield(v) {
			return
		}
		v, ok := pulls[v.index()]()
		if ok {
			heap.Push(x, v)
		}
	}
}
