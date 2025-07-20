package kway

import (
	"cmp"
	"iter"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Helper function to collect all values from an iter.Seq
func collectSeq[T any](seq iter.Seq[T]) []T {
	var result []T
	for v := range seq {
		result = append(result, v)
	}
	return result
}

// Helper function to collect all values from an iter.Seq2
func collectSeq2[T1, T2 any](seq iter.Seq2[T1, T2]) ([]T1, []T2) {
	var r1 []T1
	var r2 []T2
	for v1, v2 := range seq {
		r1 = append(r1, v1)
		r2 = append(r2, v2)
	}
	return r1, r2
}

// Helper function to create an iter.Seq from a slice
func sliceSeq[T any](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Helper function to create an iter.Seq2 from two slices
func sliceSeq2[T1, T2 any](s1 []T1, s2 []T2) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for i := 0; i < len(s1) && i < len(s2); i++ {
			if !yield(s1[i], s2[i]) {
				return
			}
		}
	}
}

func TestMerge_NilCompareFunction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil comparison function")
		} else if !strings.Contains(r.(string), "nil comparison function") {
			t.Errorf("Expected panic message about nil comparison function, got: %v", r)
		}
	}()

	seq := sliceSeq([]int{1, 2, 3})
	_ = Merge[int](nil, seq)
}

func TestMerge_EmptyInput(t *testing.T) {
	// No sequences
	result := collectSeq(Merge(cmp.Compare[int]))
	if len(result) != 0 {
		t.Errorf("Expected empty result for no sequences, got %v", result)
	}

	// All nil sequences
	result = collectSeq(Merge(cmp.Compare[int], nil, nil, nil))
	if len(result) != 0 {
		t.Errorf("Expected empty result for all nil sequences, got %v", result)
	}

	// Mix of nil and empty sequences
	empty := sliceSeq([]int{})
	result = collectSeq(Merge(cmp.Compare[int], nil, empty, nil))
	if len(result) != 0 {
		t.Errorf("Expected empty result for nil and empty sequences, got %v", result)
	}
}

func TestMerge_EmptySeqFunction(t *testing.T) {
	// Test the emptySeq function directly by triggering it via edge cases
	// This happens when all sequences are nil

	// Test with completely empty input (no sequences at all)
	var result []int
	for v := range Merge(cmp.Compare[int]) {
		result = append(result, v)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}

	// Test with all nil sequences - this should trigger emptySeq
	var nilSeqs []iter.Seq[int]
	nilSeqs = append(nilSeqs, nil, nil, nil)
	result = collectSeq(Merge(cmp.Compare[int], nilSeqs...))
	if len(result) != 0 {
		t.Errorf("Expected empty result for all nil sequences, got %v", result)
	}

	// Test early termination of empty sequence
	count := 0
	for range Merge(cmp.Compare[int]) {
		count++
		if count > 0 { // should never happen
			break
		}
	}
	if count != 0 {
		t.Errorf("Expected no iterations for empty sequence, got %d", count)
	}
}

func TestMerge_SingleSequence(t *testing.T) {
	// Single non-empty sequence
	input := []int{1, 3, 5, 7}
	result := collectSeq(Merge(cmp.Compare[int], sliceSeq(input)))
	if !slices.Equal(result, input) {
		t.Errorf("Expected %v, got %v", input, result)
	}

	// Single empty sequence
	result = collectSeq(Merge(cmp.Compare[int], sliceSeq([]int{})))
	if len(result) != 0 {
		t.Errorf("Expected empty result for single empty sequence, got %v", result)
	}
}

func TestMerge_TwoSequences(t *testing.T) {
	tests := []struct {
		name     string
		seq1     []int
		seq2     []int
		expected []int
	}{
		{
			name:     "both sequences same length",
			seq1:     []int{1, 3, 5},
			seq2:     []int{2, 4, 6},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "first sequence longer",
			seq1:     []int{1, 3, 5, 7, 9},
			seq2:     []int{2, 4},
			expected: []int{1, 2, 3, 4, 5, 7, 9},
		},
		{
			name:     "second sequence longer",
			seq1:     []int{1, 3},
			seq2:     []int{2, 4, 6, 8, 10},
			expected: []int{1, 2, 3, 4, 6, 8, 10},
		},
		{
			name:     "one sequence empty",
			seq1:     []int{},
			seq2:     []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "overlapping values",
			seq1:     []int{1, 3, 5},
			seq2:     []int{1, 3, 5},
			expected: []int{1, 1, 3, 3, 5, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectSeq(Merge(cmp.Compare[int], sliceSeq(tt.seq1), sliceSeq(tt.seq2)))
			if !slices.Equal(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMerge_MultipleSequences(t *testing.T) {
	seq1 := []int{1, 5, 9}
	seq2 := []int{2, 6, 10}
	seq3 := []int{3, 7, 11}
	seq4 := []int{4, 8, 12}
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	result := collectSeq(Merge(cmp.Compare[int],
		sliceSeq(seq1),
		sliceSeq(seq2),
		sliceSeq(seq3),
		sliceSeq(seq4),
	))

	if !slices.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMerge_Stability(t *testing.T) {
	// Test stability: when comparison returns 0, order should be preserved by sequence index
	type stableValue struct {
		value int
		seqID int // which sequence this came from
	}

	cmpFunc := func(a, b stableValue) int {
		return cmp.Compare(a.value, b.value)
	}

	seq1 := sliceSeq([]stableValue{{1, 1}, {2, 1}, {3, 1}})
	seq2 := sliceSeq([]stableValue{{1, 2}, {2, 2}, {3, 2}})
	seq3 := sliceSeq([]stableValue{{1, 3}, {2, 3}, {3, 3}})

	result := collectSeq(Merge(cmpFunc, seq1, seq2, seq3))

	// Check that for each value, sequence order is preserved
	expected := []stableValue{
		{1, 1}, {1, 2}, {1, 3}, // all 1s in sequence order
		{2, 1}, {2, 2}, {2, 3}, // all 2s in sequence order
		{3, 1}, {3, 2}, {3, 3}, // all 3s in sequence order
	}

	if !slices.Equal(result, expected) {
		t.Errorf("Stability test failed. Expected %v, got %v", expected, result)
	}
}

func TestMerge_StringComparison(t *testing.T) {
	seq1 := sliceSeq([]string{"apple", "cherry", "grape"})
	seq2 := sliceSeq([]string{"banana", "elderberry", "fig"})
	expected := []string{"apple", "banana", "cherry", "elderberry", "fig", "grape"}

	result := collectSeq(Merge(strings.Compare, seq1, seq2))

	if !slices.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMerge_ReverseOrder(t *testing.T) {
	reverseCompare := func(a, b int) int {
		return cmp.Compare(b, a) // reverse order
	}

	seq1 := sliceSeq([]int{9, 5, 1}) // descending
	seq2 := sliceSeq([]int{8, 4, 2}) // descending
	expected := []int{9, 8, 5, 4, 2, 1}

	result := collectSeq(Merge(reverseCompare, seq1, seq2))

	if !slices.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMerge_EarlyTermination(t *testing.T) {
	seq1 := sliceSeq([]int{1, 3, 5, 7, 9})
	seq2 := sliceSeq([]int{2, 4, 6, 8, 10})

	var result []int
	for v := range Merge(cmp.Compare[int], seq1, seq2) {
		result = append(result, v)
		if len(result) == 3 { // terminate early
			break
		}
	}

	expected := []int{1, 2, 3}
	if !slices.Equal(result, expected) {
		t.Errorf("Early termination test failed. Expected %v, got %v", expected, result)
	}
}

func TestMerge_LargeSequences(t *testing.T) {
	// Test with larger sequences to verify performance and correctness
	size := 1000
	seq1 := make([]int, size)
	seq2 := make([]int, size)

	for i := 0; i < size; i++ {
		seq1[i] = i * 2   // even numbers
		seq2[i] = i*2 + 1 // odd numbers
	}

	result := collectSeq(Merge(cmp.Compare[int], sliceSeq(seq1), sliceSeq(seq2)))

	// Verify length
	if len(result) != size*2 {
		t.Errorf("Expected length %d, got %d", size*2, len(result))
	}

	// Verify sorted order
	for i := 1; i < len(result); i++ {
		if result[i] < result[i-1] {
			t.Errorf("Result not sorted at index %d: %d < %d", i, result[i], result[i-1])
		}
	}

	// Verify all numbers 0 to 2*size-1 are present
	for i := 0; i < size*2; i++ {
		if result[i] != i {
			t.Errorf("Expected %d at index %d, got %d", i, i, result[i])
		}
	}
}

// Test Merge2 function

func TestMerge2_NilCompareFunction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil comparison function")
		} else if !strings.Contains(r.(string), "nil comparison function") {
			t.Errorf("Expected panic message about nil comparison function, got: %v", r)
		}
	}()

	seq := sliceSeq2([]int{1, 2}, []string{"a", "b"})
	_ = Merge2[int, string](nil, seq)
}

func TestMerge2_EmptyInput(t *testing.T) {
	cmpFunc := func(a1 int, a2 string, b1 int, b2 string) int {
		return cmp.Compare(a1, b1)
	}

	// No sequences
	r1, r2 := collectSeq2(Merge2(cmpFunc))
	if len(r1) != 0 || len(r2) != 0 {
		t.Errorf("Expected empty result for no sequences, got %v, %v", r1, r2)
	}

	// All nil sequences
	r1, r2 = collectSeq2(Merge2(cmpFunc, nil, nil))
	if len(r1) != 0 || len(r2) != 0 {
		t.Errorf("Expected empty result for all nil sequences, got %v, %v", r1, r2)
	}
}

func TestMerge2_EmptySeq2Function(t *testing.T) {
	cmpFunc := func(a1 int, a2 string, b1 int, b2 string) int {
		return cmp.Compare(a1, b1)
	}

	// Test the emptySeq2 function directly by triggering it via edge cases
	// This happens when all sequences are nil

	// Test with completely empty input (no sequences at all)
	var r1 []int
	var r2 []string
	for v1, v2 := range Merge2(cmpFunc) {
		r1 = append(r1, v1)
		r2 = append(r2, v2)
	}
	if len(r1) != 0 || len(r2) != 0 {
		t.Errorf("Expected empty result, got %v, %v", r1, r2)
	}

	// Test with all nil sequences - this should trigger emptySeq2
	var nilSeqs []iter.Seq2[int, string]
	nilSeqs = append(nilSeqs, nil, nil, nil)
	r1, r2 = collectSeq2(Merge2(cmpFunc, nilSeqs...))
	if len(r1) != 0 || len(r2) != 0 {
		t.Errorf("Expected empty result for all nil sequences, got %v, %v", r1, r2)
	}

	// Test early termination of empty sequence
	count := 0
	for range Merge2(cmpFunc) {
		count++
		if count > 0 { // should never happen
			break
		}
	}
	if count != 0 {
		t.Errorf("Expected no iterations for empty sequence, got %d", count)
	}
}

func TestMerge2_SingleSequence(t *testing.T) {
	cmpFunc := func(a1 int, a2 string, b1 int, b2 string) int {
		return cmp.Compare(a1, b1)
	}

	input1 := []int{1, 3, 5}
	input2 := []string{"a", "c", "e"}
	r1, r2 := collectSeq2(Merge2(cmpFunc, sliceSeq2(input1, input2)))

	if !slices.Equal(r1, input1) || !slices.Equal(r2, input2) {
		t.Errorf("Expected %v, %v; got %v, %v", input1, input2, r1, r2)
	}
}

func TestMerge2_TwoSequences(t *testing.T) {
	cmpFunc := func(a1 int, a2 string, b1 int, b2 string) int {
		return cmp.Compare(a1, b1)
	}

	seq1_1 := []int{1, 5, 9}
	seq1_2 := []string{"a", "e", "i"}
	seq2_1 := []int{3, 7, 11}
	seq2_2 := []string{"c", "g", "k"}

	r1, r2 := collectSeq2(Merge2(cmpFunc,
		sliceSeq2(seq1_1, seq1_2),
		sliceSeq2(seq2_1, seq2_2),
	))

	expected1 := []int{1, 3, 5, 7, 9, 11}
	expected2 := []string{"a", "c", "e", "g", "i", "k"}

	if !slices.Equal(r1, expected1) || !slices.Equal(r2, expected2) {
		t.Errorf("Expected %v, %v; got %v, %v", expected1, expected2, r1, r2)
	}
}

func TestMerge2_ComplexComparison(t *testing.T) {
	// Compare by string length first, then alphabetically
	cmpFunc := func(a1 int, a2 string, b1 int, b2 string) int {
		if lenCmp := cmp.Compare(len(a2), len(b2)); lenCmp != 0 {
			return lenCmp
		}
		return strings.Compare(a2, b2)
	}

	seq1_1 := []int{1, 2, 3}
	seq1_2 := []string{"a", "bb", "ccc"}
	seq2_1 := []int{4, 5, 6}
	seq2_2 := []string{"d", "ee", "fff"}

	r1, r2 := collectSeq2(Merge2(cmpFunc,
		sliceSeq2(seq1_1, seq1_2),
		sliceSeq2(seq2_1, seq2_2),
	))

	expected1 := []int{1, 4, 2, 5, 3, 6}
	expected2 := []string{"a", "d", "bb", "ee", "ccc", "fff"}

	if !slices.Equal(r1, expected1) || !slices.Equal(r2, expected2) {
		t.Errorf("Expected %v, %v; got %v, %v", expected1, expected2, r1, r2)
	}
}

func TestMerge2_EarlyTermination(t *testing.T) {
	cmpFunc := func(a1 int, a2 string, b1 int, b2 string) int {
		return cmp.Compare(a1, b1)
	}

	seq1_1 := []int{1, 5, 9}
	seq1_2 := []string{"a", "e", "i"}
	seq2_1 := []int{3, 7, 11}
	seq2_2 := []string{"c", "g", "k"}

	var r1, r2 []string
	count := 0
	for v1, v2 := range Merge2(cmpFunc,
		sliceSeq2(seq1_1, seq1_2),
		sliceSeq2(seq2_1, seq2_2),
	) {
		r1 = append(r1, strconv.Itoa(v1))
		r2 = append(r2, v2)
		count++
		if count == 3 { // terminate early
			break
		}
	}

	expected1 := []string{"1", "3", "5"}
	expected2 := []string{"a", "c", "e"}

	if !slices.Equal(r1, expected1) || !slices.Equal(r2, expected2) {
		t.Errorf("Early termination test failed. Expected %v, %v; got %v, %v", expected1, expected2, r1, r2)
	}
}

func TestMerge2_Stability(t *testing.T) {
	type stablePair struct {
		value int
		seqID int
	}

	cmpFunc := func(a1 stablePair, a2 string, b1 stablePair, b2 string) int {
		return cmp.Compare(a1.value, b1.value)
	}

	seq1_1 := []stablePair{{1, 1}, {2, 1}}
	seq1_2 := []string{"a1", "b1"}
	seq2_1 := []stablePair{{1, 2}, {2, 2}}
	seq2_2 := []string{"a2", "b2"}

	r1, r2 := collectSeq2(Merge2(cmpFunc,
		sliceSeq2(seq1_1, seq1_2),
		sliceSeq2(seq2_1, seq2_2),
	))

	expected1 := []stablePair{{1, 1}, {1, 2}, {2, 1}, {2, 2}}
	expected2 := []string{"a1", "a2", "b1", "b2"}

	if !slices.Equal(r1, expected1) || !slices.Equal(r2, expected2) {
		t.Errorf("Stability test failed. Expected %v, %v; got %v, %v", expected1, expected2, r1, r2)
	}
}

func BenchmarkMerge_TwoSequences(b *testing.B) {
	seq1 := make([]int, 1000)
	seq2 := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		seq1[i] = i * 2
		seq2[i] = i*2 + 1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := collectSeq(Merge(cmp.Compare[int], sliceSeq(seq1), sliceSeq(seq2)))
		_ = result
	}
}

func BenchmarkMerge_MultipleSequences(b *testing.B) {
	seqs := make([]iter.Seq[int], 10)
	for i := 0; i < 10; i++ {
		seq := make([]int, 100)
		for j := 0; j < 100; j++ {
			seq[j] = i + j*10
		}
		seqs[i] = sliceSeq(seq)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := collectSeq(Merge(cmp.Compare[int], seqs...))
		_ = result
	}
}

func Test_emptySeq(t *testing.T) {
	emptySeq[int](func(v int) bool {
		t.Fatal("unexpected call to emptySeq")
		return true
	})
}

func Test_emptySeq2(t *testing.T) {
	emptySeq2[int16, int32](func(a int16, b int32) bool {
		t.Fatal("unexpected call to emptySeq")
		return true
	})
}
