package it

import "iter"

// ForIndex for each index in the slice
func ForIndex[T any](items []T, fn func(int)) {
	for i := range items {
		fn(i)
	}
}

// Range for each item in the slice
func Range[T any](items []T, fn func(T)) {
	for _, item := range items {
		fn(item)
	}
}

// Each each item in the slice
func Each[T any](items []T, fn func(int, T)) {
	for i, item := range items {
		fn(i, item)
	}
}

// Iter for each item in the iterator
func Iter[T any](iter iter.Seq[T], fn func(T)) {
	for item := range iter {
		fn(item)
	}
}

// Iter2 for each item in the iterator
func Iter2[T any](iter iter.Seq2[int, T], fn func(int, T)) {
	for i, item := range iter {
		fn(i, item)
	}
}
