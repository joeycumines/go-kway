package kway

import "iter"

type wrappedSeqValue[T any] struct {
	i int
	v T
}

func (x *wrappedSeqValue[T]) index() int { return x.i }

type wrappedSeq2Value[T1 any, T2 any] struct {
	i  int
	v1 T1
	v2 T2
}

func (x *wrappedSeq2Value[T1, T2]) index() int { return x.i }

func wrapSeq[T any](i int, seq iter.Seq[T]) iter.Seq[*wrappedSeqValue[T]] {
	return func(yield func(*wrappedSeqValue[T]) bool) {
		for v := range seq {
			if !yield(&wrappedSeqValue[T]{i, v}) {
				return
			}
		}
	}
}

func wrapSeq2[T1 any, T2 any](i int, seq iter.Seq2[T1, T2]) iter.Seq[*wrappedSeq2Value[T1, T2]] {
	return func(yield func(*wrappedSeq2Value[T1, T2]) bool) {
		for v1, v2 := range seq {
			if !yield(&wrappedSeq2Value[T1, T2]{i, v1, v2}) {
				return
			}
		}
	}
}

func wrapCompare[T any](compare func(a, b T) int) func(a, b *wrappedSeqValue[T]) int {
	return func(a, b *wrappedSeqValue[T]) int {
		return compare(a.v, b.v)
	}
}

func wrapCompare2[T1 any, T2 any](compare func(a1 T1, a2 T2, b1 T1, b2 T2) int) func(a, b *wrappedSeq2Value[T1, T2]) int {
	return func(a, b *wrappedSeq2Value[T1, T2]) int {
		return compare(a.v1, a.v2, b.v1, b.v2)
	}
}
