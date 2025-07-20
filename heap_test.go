package kway

import (
	"cmp"
	"container/heap"
	"iter"
	"slices"
	"testing"
)

// Mock implementation of index() interface for testing
type mockIndexValue struct {
	value int
	idx   int
}

func (m *mockIndexValue) index() int {
	return m.idx
}

func TestMergeState_Len(t *testing.T) {
	ms := &mergeState[*mockIndexValue]{
		items: []*mockIndexValue{
			{value: 1, idx: 0},
			{value: 2, idx: 1},
			{value: 3, idx: 2},
		},
	}

	if ms.Len() != 3 {
		t.Errorf("Expected Len() = 3, got %d", ms.Len())
	}

	ms.items = nil
	if ms.Len() != 0 {
		t.Errorf("Expected Len() = 0 for nil items, got %d", ms.Len())
	}
}

func TestMergeState_Less(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	ms := &mergeState[*mockIndexValue]{
		cmp: cmpFunc,
		items: []*mockIndexValue{
			{value: 2, idx: 1}, // index 0
			{value: 1, idx: 0}, // index 1
			{value: 3, idx: 2}, // index 2
		},
	}

	// Test comparison by value (2 > 1)
	if ms.Less(0, 1) {
		t.Error("Expected items[0] (value=2) NOT less than items[1] (value=1)")
	}

	// Test comparison by value (1 < 2)
	if !ms.Less(1, 0) {
		t.Error("Expected items[1] (value=1) less than items[0] (value=2)")
	}

	// Test tiebreaker by index when values are equal
	ms.items = []*mockIndexValue{
		{value: 5, idx: 2}, // index 0
		{value: 5, idx: 1}, // index 1
	}

	if ms.Less(0, 1) {
		t.Error("Expected items[0] (idx=2) NOT less than items[1] (idx=1) when values equal")
	}

	if !ms.Less(1, 0) {
		t.Error("Expected items[1] (idx=1) less than items[0] (idx=2) when values equal")
	}
}

func TestMergeState_Swap(t *testing.T) {
	ms := &mergeState[*mockIndexValue]{
		items: []*mockIndexValue{
			{value: 1, idx: 0},
			{value: 2, idx: 1},
			{value: 3, idx: 2},
		},
	}

	originalFirst := ms.items[0]
	originalSecond := ms.items[1]

	ms.Swap(0, 1)

	if ms.items[0] != originalSecond {
		t.Error("Expected items[0] to be swapped with original items[1]")
	}
	if ms.items[1] != originalFirst {
		t.Error("Expected items[1] to be swapped with original items[0]")
	}
	if ms.items[2].value != 3 {
		t.Error("Expected items[2] to remain unchanged")
	}
}

func TestMergeState_Push(t *testing.T) {
	ms := &mergeState[*mockIndexValue]{
		items: []*mockIndexValue{
			{value: 1, idx: 0},
		},
	}

	newItem := &mockIndexValue{value: 2, idx: 1}
	ms.Push(newItem)

	if len(ms.items) != 2 {
		t.Errorf("Expected length 2 after Push, got %d", len(ms.items))
	}
	if ms.items[1] != newItem {
		t.Error("Expected new item to be added at the end")
	}
}

func TestMergeState_Pop(t *testing.T) {
	ms := &mergeState[*mockIndexValue]{
		items: []*mockIndexValue{
			{value: 1, idx: 0},
			{value: 2, idx: 1},
			{value: 3, idx: 2},
		},
	}

	popped := ms.Pop().(*mockIndexValue)

	if len(ms.items) != 2 {
		t.Errorf("Expected length 2 after Pop, got %d", len(ms.items))
	}
	if popped.value != 3 || popped.idx != 2 {
		t.Errorf("Expected popped item to be {value: 3, idx: 2}, got {value: %d, idx: %d}", popped.value, popped.idx)
	}

	// Verify the popped location is zeroed
	// Note: We can't easily test this directly without accessing internal implementation details
	// but the Pop method should zero out the removed element
}

func TestMergeState_HeapInterface(t *testing.T) {
	// Test that mergeState properly implements heap.Interface
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	ms := &mergeState[*mockIndexValue]{
		cmp: cmpFunc,
		items: []*mockIndexValue{
			{value: 5, idx: 0},
			{value: 2, idx: 1},
			{value: 8, idx: 2},
			{value: 1, idx: 3},
			{value: 6, idx: 4},
		},
	}

	// Initialize heap
	heap.Init(ms)

	// Verify heap property is maintained
	for len(ms.items) > 0 {
		min := heap.Pop(ms).(*mockIndexValue)
		// Next element should be >= current minimum
		if len(ms.items) > 0 {
			next := ms.items[0]
			if ms.cmp(next, min) < 0 {
				t.Errorf("Heap property violated: next item %v < popped item %v", next, min)
			}
		}
	}
}

func TestMergeState_All_EmptySequences(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	ms := &mergeState[*mockIndexValue]{
		cmp:  cmpFunc,
		seqs: []iter.Seq[*mockIndexValue]{},
	}

	var result []*mockIndexValue
	for v := range ms.all {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty sequences, got %v", result)
	}
}

func TestMergeState_All_NilSequences(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	ms := &mergeState[*mockIndexValue]{
		cmp:  cmpFunc,
		seqs: []iter.Seq[*mockIndexValue]{nil, nil, nil},
	}

	var result []*mockIndexValue
	for v := range ms.all {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for nil sequences, got %v", result)
	}
}

func TestMergeState_All_SingleSequence(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	input := []*mockIndexValue{
		{value: 1, idx: 0},
		{value: 3, idx: 0},
		{value: 5, idx: 0},
	}

	seq := func(yield func(*mockIndexValue) bool) {
		for _, v := range input {
			if !yield(v) {
				return
			}
		}
	}

	ms := &mergeState[*mockIndexValue]{
		cmp:  cmpFunc,
		seqs: []iter.Seq[*mockIndexValue]{seq},
	}

	var result []*mockIndexValue
	for v := range ms.all {
		result = append(result, v)
	}

	if !slices.EqualFunc(result, input, func(a, b *mockIndexValue) bool {
		return a.value == b.value && a.idx == b.idx
	}) {
		t.Errorf("Expected %v, got %v", input, result)
	}
}

func TestMergeState_All_MultipleSequences(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	seq1 := []*mockIndexValue{
		{value: 1, idx: 0},
		{value: 4, idx: 0},
		{value: 7, idx: 0},
	}

	seq2 := []*mockIndexValue{
		{value: 2, idx: 1},
		{value: 5, idx: 1},
		{value: 8, idx: 1},
	}

	seq3 := []*mockIndexValue{
		{value: 3, idx: 2},
		{value: 6, idx: 2},
		{value: 9, idx: 2},
	}

	seqFunc1 := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq1 {
			if !yield(v) {
				return
			}
		}
	}

	seqFunc2 := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq2 {
			if !yield(v) {
				return
			}
		}
	}

	seqFunc3 := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq3 {
			if !yield(v) {
				return
			}
		}
	}

	ms := &mergeState[*mockIndexValue]{
		cmp:  cmpFunc,
		seqs: []iter.Seq[*mockIndexValue]{seqFunc1, seqFunc2, seqFunc3},
	}

	var result []*mockIndexValue
	for v := range ms.all {
		result = append(result, v)
	}

	expected := []*mockIndexValue{
		{value: 1, idx: 0},
		{value: 2, idx: 1},
		{value: 3, idx: 2},
		{value: 4, idx: 0},
		{value: 5, idx: 1},
		{value: 6, idx: 2},
		{value: 7, idx: 0},
		{value: 8, idx: 1},
		{value: 9, idx: 2},
	}

	if !slices.EqualFunc(result, expected, func(a, b *mockIndexValue) bool {
		return a.value == b.value && a.idx == b.idx
	}) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMergeState_All_EarlyTermination(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	seq1 := []*mockIndexValue{
		{value: 1, idx: 0},
		{value: 4, idx: 0},
		{value: 7, idx: 0},
	}

	seq2 := []*mockIndexValue{
		{value: 2, idx: 1},
		{value: 5, idx: 1},
		{value: 8, idx: 1},
	}

	seqFunc1 := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq1 {
			if !yield(v) {
				return
			}
		}
	}

	seqFunc2 := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq2 {
			if !yield(v) {
				return
			}
		}
	}

	ms := &mergeState[*mockIndexValue]{
		cmp:  cmpFunc,
		seqs: []iter.Seq[*mockIndexValue]{seqFunc1, seqFunc2},
	}

	var result []*mockIndexValue
	count := 0
	for v := range ms.all {
		result = append(result, v)
		count++
		if count == 3 { // terminate early
			break
		}
	}

	expected := []*mockIndexValue{
		{value: 1, idx: 0},
		{value: 2, idx: 1},
		{value: 4, idx: 0},
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}

	if !slices.EqualFunc(result, expected, func(a, b *mockIndexValue) bool {
		return a.value == b.value && a.idx == b.idx
	}) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMergeState_All_StabilityPreservation(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	// Two sequences with identical values - should preserve sequence order
	seq1 := []*mockIndexValue{
		{value: 1, idx: 0},
		{value: 2, idx: 0},
		{value: 3, idx: 0},
	}

	seq2 := []*mockIndexValue{
		{value: 1, idx: 1},
		{value: 2, idx: 1},
		{value: 3, idx: 1},
	}

	seqFunc1 := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq1 {
			if !yield(v) {
				return
			}
		}
	}

	seqFunc2 := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq2 {
			if !yield(v) {
				return
			}
		}
	}

	ms := &mergeState[*mockIndexValue]{
		cmp:  cmpFunc,
		seqs: []iter.Seq[*mockIndexValue]{seqFunc1, seqFunc2},
	}

	var result []*mockIndexValue
	for v := range ms.all {
		result = append(result, v)
	}

	// Verify that for each value, sequence order is preserved (lower index first)
	expected := []*mockIndexValue{
		{value: 1, idx: 0}, // seq 0 comes first
		{value: 1, idx: 1}, // seq 1 comes second
		{value: 2, idx: 0}, // seq 0 comes first
		{value: 2, idx: 1}, // seq 1 comes second
		{value: 3, idx: 0}, // seq 0 comes first
		{value: 3, idx: 1}, // seq 1 comes second
	}

	if !slices.EqualFunc(result, expected, func(a, b *mockIndexValue) bool {
		return a.value == b.value && a.idx == b.idx
	}) {
		t.Errorf("Stability test failed. Expected %v, got %v", expected, result)
	}
}

func TestMergeState_All_MixedNilAndValidSequences(t *testing.T) {
	cmpFunc := func(a, b *mockIndexValue) int {
		return cmp.Compare(a.value, b.value)
	}

	seq := []*mockIndexValue{
		{value: 1, idx: 1},
		{value: 3, idx: 1},
		{value: 5, idx: 1},
	}

	seqFunc := func(yield func(*mockIndexValue) bool) {
		for _, v := range seq {
			if !yield(v) {
				return
			}
		}
	}

	ms := &mergeState[*mockIndexValue]{
		cmp:  cmpFunc,
		seqs: []iter.Seq[*mockIndexValue]{nil, seqFunc, nil},
	}

	var result []*mockIndexValue
	for v := range ms.all {
		result = append(result, v)
	}

	if !slices.EqualFunc(result, seq, func(a, b *mockIndexValue) bool {
		return a.value == b.value && a.idx == b.idx
	}) {
		t.Errorf("Expected %v, got %v", seq, result)
	}
}
