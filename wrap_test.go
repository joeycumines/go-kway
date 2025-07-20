package kway

import (
	"cmp"
	"slices"
	"strings"
	"testing"
)

func TestWrappedSeqValue_index(t *testing.T) {
	tests := []struct {
		name     string
		value    *wrappedSeqValue[int]
		expected int
	}{
		{
			name:     "zero index",
			value:    &wrappedSeqValue[int]{i: 0, v: 42},
			expected: 0,
		},
		{
			name:     "positive index",
			value:    &wrappedSeqValue[int]{i: 5, v: 42},
			expected: 5,
		},
		{
			name:     "large index",
			value:    &wrappedSeqValue[int]{i: 1000, v: 42},
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.index(); got != tt.expected {
				t.Errorf("index() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestWrappedSeq2Value_index(t *testing.T) {
	tests := []struct {
		name     string
		value    *wrappedSeq2Value[int, string]
		expected int
	}{
		{
			name:     "zero index",
			value:    &wrappedSeq2Value[int, string]{i: 0, v1: 42, v2: "test"},
			expected: 0,
		},
		{
			name:     "positive index",
			value:    &wrappedSeq2Value[int, string]{i: 7, v1: 42, v2: "test"},
			expected: 7,
		},
		{
			name:     "large index",
			value:    &wrappedSeq2Value[int, string]{i: 2000, v1: 42, v2: "test"},
			expected: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.index(); got != tt.expected {
				t.Errorf("index() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestWrapSeq(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		input    []int
		expected []*wrappedSeqValue[int]
	}{
		{
			name:     "empty sequence",
			index:    0,
			input:    []int{},
			expected: []*wrappedSeqValue[int]{},
		},
		{
			name:  "single element",
			index: 2,
			input: []int{42},
			expected: []*wrappedSeqValue[int]{
				{i: 2, v: 42},
			},
		},
		{
			name:  "multiple elements",
			index: 1,
			input: []int{1, 2, 3, 4, 5},
			expected: []*wrappedSeqValue[int]{
				{i: 1, v: 1},
				{i: 1, v: 2},
				{i: 1, v: 3},
				{i: 1, v: 4},
				{i: 1, v: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := sliceSeq(tt.input)
			wrappedSeq := wrapSeq(tt.index, seq)

			var result []*wrappedSeqValue[int]
			for v := range wrappedSeq {
				result = append(result, v)
			}

			if !slices.EqualFunc(result, tt.expected, func(a, b *wrappedSeqValue[int]) bool {
				return a.i == b.i && a.v == b.v
			}) {
				t.Errorf("wrapSeq() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWrapSeq_EarlyTermination(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	seq := sliceSeq(input)
	wrappedSeq := wrapSeq(0, seq)

	var result []*wrappedSeqValue[int]
	count := 0
	for v := range wrappedSeq {
		result = append(result, v)
		count++
		if count == 3 { // terminate early
			break
		}
	}

	expected := []*wrappedSeqValue[int]{
		{i: 0, v: 1},
		{i: 0, v: 2},
		{i: 0, v: 3},
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}

	if !slices.EqualFunc(result, expected, func(a, b *wrappedSeqValue[int]) bool {
		return a.i == b.i && a.v == b.v
	}) {
		t.Errorf("Early termination test failed. Expected %v, got %v", expected, result)
	}
}

func TestWrapSeq2(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		input1   []int
		input2   []string
		expected []*wrappedSeq2Value[int, string]
	}{
		{
			name:     "empty sequence",
			index:    0,
			input1:   []int{},
			input2:   []string{},
			expected: []*wrappedSeq2Value[int, string]{},
		},
		{
			name:   "single element",
			index:  3,
			input1: []int{42},
			input2: []string{"test"},
			expected: []*wrappedSeq2Value[int, string]{
				{i: 3, v1: 42, v2: "test"},
			},
		},
		{
			name:   "multiple elements",
			index:  2,
			input1: []int{1, 2, 3},
			input2: []string{"a", "b", "c"},
			expected: []*wrappedSeq2Value[int, string]{
				{i: 2, v1: 1, v2: "a"},
				{i: 2, v1: 2, v2: "b"},
				{i: 2, v1: 3, v2: "c"},
			},
		},
		{
			name:   "mismatched lengths",
			index:  1,
			input1: []int{1, 2},
			input2: []string{"a", "b", "c", "d"},
			expected: []*wrappedSeq2Value[int, string]{
				{i: 1, v1: 1, v2: "a"},
				{i: 1, v1: 2, v2: "b"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := sliceSeq2(tt.input1, tt.input2)
			wrappedSeq := wrapSeq2(tt.index, seq)

			var result []*wrappedSeq2Value[int, string]
			for v := range wrappedSeq {
				result = append(result, v)
			}

			if !slices.EqualFunc(result, tt.expected, func(a, b *wrappedSeq2Value[int, string]) bool {
				return a.i == b.i && a.v1 == b.v1 && a.v2 == b.v2
			}) {
				t.Errorf("wrapSeq2() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWrapSeq2_EarlyTermination(t *testing.T) {
	input1 := []int{1, 2, 3, 4, 5}
	input2 := []string{"a", "b", "c", "d", "e"}
	seq := sliceSeq2(input1, input2)
	wrappedSeq := wrapSeq2(1, seq)

	var result []*wrappedSeq2Value[int, string]
	count := 0
	for v := range wrappedSeq {
		result = append(result, v)
		count++
		if count == 2 { // terminate early
			break
		}
	}

	expected := []*wrappedSeq2Value[int, string]{
		{i: 1, v1: 1, v2: "a"},
		{i: 1, v1: 2, v2: "b"},
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}

	if !slices.EqualFunc(result, expected, func(a, b *wrappedSeq2Value[int, string]) bool {
		return a.i == b.i && a.v1 == b.v1 && a.v2 == b.v2
	}) {
		t.Errorf("Early termination test failed. Expected %v, got %v", expected, result)
	}
}

func TestWrapCompare(t *testing.T) {
	originalCompare := cmp.Compare[int]
	wrappedCompare := wrapCompare(originalCompare)

	tests := []struct {
		name     string
		a        *wrappedSeqValue[int]
		b        *wrappedSeqValue[int]
		expected int
	}{
		{
			name:     "a less than b",
			a:        &wrappedSeqValue[int]{i: 0, v: 1},
			b:        &wrappedSeqValue[int]{i: 1, v: 2},
			expected: -1,
		},
		{
			name:     "a equal to b",
			a:        &wrappedSeqValue[int]{i: 0, v: 5},
			b:        &wrappedSeqValue[int]{i: 1, v: 5},
			expected: 0,
		},
		{
			name:     "a greater than b",
			a:        &wrappedSeqValue[int]{i: 0, v: 10},
			b:        &wrappedSeqValue[int]{i: 1, v: 3},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrappedCompare(tt.a, tt.b)
			if (result < 0 && tt.expected >= 0) ||
				(result == 0 && tt.expected != 0) ||
				(result > 0 && tt.expected <= 0) {
				t.Errorf("wrapCompare() = %d, want sign of %d", result, tt.expected)
			}
		})
	}
}

func TestWrapCompare_StringValues(t *testing.T) {
	originalCompare := strings.Compare
	wrappedCompare := wrapCompare(originalCompare)

	tests := []struct {
		name     string
		a        *wrappedSeqValue[string]
		b        *wrappedSeqValue[string]
		expected int
	}{
		{
			name:     "a less than b",
			a:        &wrappedSeqValue[string]{i: 0, v: "apple"},
			b:        &wrappedSeqValue[string]{i: 1, v: "banana"},
			expected: -1,
		},
		{
			name:     "a equal to b",
			a:        &wrappedSeqValue[string]{i: 0, v: "test"},
			b:        &wrappedSeqValue[string]{i: 1, v: "test"},
			expected: 0,
		},
		{
			name:     "a greater than b",
			a:        &wrappedSeqValue[string]{i: 0, v: "zebra"},
			b:        &wrappedSeqValue[string]{i: 1, v: "apple"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrappedCompare(tt.a, tt.b)
			if (result < 0 && tt.expected >= 0) ||
				(result == 0 && tt.expected != 0) ||
				(result > 0 && tt.expected <= 0) {
				t.Errorf("wrapCompare() = %d, want sign of %d", result, tt.expected)
			}
		})
	}
}

func TestWrapCompare2(t *testing.T) {
	originalCompare := func(a1 int, a2 string, b1 int, b2 string) int {
		if cmp := cmp.Compare(a1, b1); cmp != 0 {
			return cmp
		}
		return strings.Compare(a2, b2)
	}
	wrappedCompare := wrapCompare2(originalCompare)

	tests := []struct {
		name     string
		a        *wrappedSeq2Value[int, string]
		b        *wrappedSeq2Value[int, string]
		expected int
	}{
		{
			name:     "a less than b by first value",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 1, v2: "z"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 2, v2: "a"},
			expected: -1,
		},
		{
			name:     "a less than b by second value",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 5, v2: "apple"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 5, v2: "banana"},
			expected: -1,
		},
		{
			name:     "a equal to b",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 5, v2: "test"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 5, v2: "test"},
			expected: 0,
		},
		{
			name:     "a greater than b by first value",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 10, v2: "a"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 3, v2: "z"},
			expected: 1,
		},
		{
			name:     "a greater than b by second value",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 5, v2: "zebra"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 5, v2: "apple"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrappedCompare(tt.a, tt.b)
			if (result < 0 && tt.expected >= 0) ||
				(result == 0 && tt.expected != 0) ||
				(result > 0 && tt.expected <= 0) {
				t.Errorf("wrapCompare2() = %d, want sign of %d", result, tt.expected)
			}
		})
	}
}

func TestWrapCompare2_ComplexComparison(t *testing.T) {
	// Compare by string length first, then by string value, then by int value
	originalCompare := func(a1 int, a2 string, b1 int, b2 string) int {
		if lenCmp := cmp.Compare(len(a2), len(b2)); lenCmp != 0 {
			return lenCmp
		}
		if strCmp := strings.Compare(a2, b2); strCmp != 0 {
			return strCmp
		}
		return cmp.Compare(a1, b1)
	}
	wrappedCompare := wrapCompare2(originalCompare)

	tests := []struct {
		name     string
		a        *wrappedSeq2Value[int, string]
		b        *wrappedSeq2Value[int, string]
		expected int
	}{
		{
			name:     "different string lengths",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 1, v2: "a"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 1, v2: "bb"},
			expected: -1,
		},
		{
			name:     "same length, different strings",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 1, v2: "aa"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 1, v2: "bb"},
			expected: -1,
		},
		{
			name:     "same strings, different ints",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 1, v2: "test"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 2, v2: "test"},
			expected: -1,
		},
		{
			name:     "completely equal",
			a:        &wrappedSeq2Value[int, string]{i: 0, v1: 5, v2: "test"},
			b:        &wrappedSeq2Value[int, string]{i: 1, v1: 5, v2: "test"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrappedCompare(tt.a, tt.b)
			if (result < 0 && tt.expected >= 0) ||
				(result == 0 && tt.expected != 0) ||
				(result > 0 && tt.expected <= 0) {
				t.Errorf("wrapCompare2() complex = %d, want sign of %d", result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkWrapSeq(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}
	seq := sliceSeq(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrappedSeq := wrapSeq(0, seq)
		var count int
		for range wrappedSeq {
			count++
		}
		_ = count
	}
}

func BenchmarkWrapSeq2(b *testing.B) {
	input1 := make([]int, 1000)
	input2 := make([]string, 1000)
	for i := range input1 {
		input1[i] = i
		input2[i] = string(rune('a' + i%26))
	}
	seq := sliceSeq2(input1, input2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrappedSeq := wrapSeq2(0, seq)
		var count int
		for range wrappedSeq {
			count++
		}
		_ = count
	}
}

func BenchmarkWrapCompare(b *testing.B) {
	wrappedCompare := wrapCompare(cmp.Compare[int])
	a := &wrappedSeqValue[int]{i: 0, v: 42}
	b_val := &wrappedSeqValue[int]{i: 1, v: 43}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrappedCompare(a, b_val)
	}
}

func BenchmarkWrapCompare2(b *testing.B) {
	cmpFunc := func(a1 int, a2 string, b1 int, b2 string) int {
		if cmp := cmp.Compare(a1, b1); cmp != 0 {
			return cmp
		}
		return strings.Compare(a2, b2)
	}
	wrappedCompare := wrapCompare2(cmpFunc)
	a := &wrappedSeq2Value[int, string]{i: 0, v1: 42, v2: "test"}
	b_val := &wrappedSeq2Value[int, string]{i: 1, v1: 43, v2: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrappedCompare(a, b_val)
	}
}
