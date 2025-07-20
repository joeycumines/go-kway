package kway

import (
	"iter"
)

// Merge performs a k-way merge of the provided sorted input sequences. It
// returns a new sequence that yields the elements from all input sequences in
// sorted order.
//
// The comparison function `cmp` should behave like [cmp.Compare] or
// [strings.Compare]: it must return a negative integer if a < b, zero
// if a == b, and a positive integer if a > b.
//
// The input sequences must each be individually sorted according to `cmp`.
// The merge is stable: if cmp(a, b) == 0, the relative order of a and b in
// the output is the same as the order of the sequences they came from in the
// input.
func Merge[T any](cmp func(a, b T) int, seqs ...iter.Seq[T]) iter.Seq[T] {
	if cmp == nil {
		panic("kway: nil comparison function")
	}
	wrappedSeqs := make([]iter.Seq[*wrappedSeqValue[T]], len(seqs))
	{
		var ok bool
		for i, seq := range seqs {
			if seq != nil {
				wrappedSeqs[i] = wrapSeq(i, seq)
				ok = true
			}
		}
		if !ok {
			return emptySeq[T]
		}
	}
	return mergeSeq(wrapCompare(cmp), wrappedSeqs)
}

func emptySeq[T any](yield func(T) bool) {}

func mergeSeq[T any](cmp func(a, b *wrappedSeqValue[T]) int, seqs []iter.Seq[*wrappedSeqValue[T]]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range (&mergeState[*wrappedSeqValue[T]]{
			cmp:  cmp,
			seqs: seqs,
		}).all {
			if !yield(v.v) {
				return
			}
		}
	}
}

// Merge2 performs a k-way merge of the provided sorted input sequences. It
// returns a new sequence that yields the elements from all input sequences in
// sorted order.
//
// See [Merge] for details on the comparison function and stability.
func Merge2[T1 any, T2 any](cmp func(a1 T1, a2 T2, b1 T1, b2 T2) int, seqs ...iter.Seq2[T1, T2]) iter.Seq2[T1, T2] {
	if cmp == nil {
		panic("kway: nil comparison function")
	}
	wrappedSeqs := make([]iter.Seq[*wrappedSeq2Value[T1, T2]], len(seqs))
	{
		var ok bool
		for i, seq := range seqs {
			if seq != nil {
				wrappedSeqs[i] = wrapSeq2(i, seq)
				ok = true
			}
		}
		if !ok {
			return emptySeq2[T1, T2]
		}
	}
	return mergeSeq2(wrapCompare2(cmp), wrappedSeqs)
}

func emptySeq2[T1 any, T2 any](yield func(T1, T2) bool) {}

func mergeSeq2[T1 any, T2 any](cmp func(a, b *wrappedSeq2Value[T1, T2]) int, seqs []iter.Seq[*wrappedSeq2Value[T1, T2]]) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for v := range (&mergeState[*wrappedSeq2Value[T1, T2]]{
			cmp:  cmp,
			seqs: seqs,
		}).all {
			if !yield(v.v1, v.v2) {
				return
			}
		}
	}
}
