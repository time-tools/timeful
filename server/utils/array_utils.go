package utils

import (
	"timeful/server/models"
)

// Returns the index of the element found using the equals function, -1 if not found
func Find[T any](arr []T, equals func(T) bool) int {
	for i, v := range arr {
		if equals(v) {
			return i
		}
	}
	return -1
}

func ArrayToSet[T comparable](arr []T) models.Set[T] {
	set := make(models.Set[T])

	for _, v := range arr {
		set[v] = struct{}{}
	}

	return set
}

type ElementWithIndex[T any] struct {
	Index int
	Value T
}

// Returns an array of elements that were added, removed, and kept based on the arrays that were passed in
func FindAddedRemovedKept[T comparable](arr []T, origArr []T) ([]ElementWithIndex[T], []ElementWithIndex[T], []ElementWithIndex[T]) {
	added := make([]ElementWithIndex[T], 0)
	removed := make([]ElementWithIndex[T], 0)
	kept := make([]ElementWithIndex[T], 0)

	// Find elements that were removed / kept
	for i, origVal := range origArr {
		// Check if original element is present in updated array
		found := false
		for _, val := range arr {
			if val == origVal {
				found = true
				break
			}
		}

		if found {
			// Element was kept
			kept = append(kept, ElementWithIndex[T]{Index: i, Value: origVal})
		} else {
			// Element was removed
			removed = append(removed, ElementWithIndex[T]{Index: i, Value: origVal})
		}
	}

	// Find elements that were added
	for i, val := range arr {
		found := false
		for _, origVal := range kept {
			if val == origVal.Value {
				found = true
				break
			}
		}

		if !found {
			// Element was added
			added = append(added, ElementWithIndex[T]{Index: i, Value: val})
		}
	}

	return added, removed, kept
}
