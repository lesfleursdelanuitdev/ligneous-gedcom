package query

import (
	"fmt"
)

// applyUniquenessByStringKey applies uniqueness logic using string keys extracted from items.
// This is a generic helper for uniqueness operations that use string keys.
func applyUniquenessByStringKey[T any](
	items []T,
	keyExtractor func(T) string,
	skipEmpty bool,
) []T {
	seen := make(map[string]bool)
	result := make([]T, 0)
	for _, item := range items {
		key := keyExtractor(item)
		if skipEmpty && key == "" {
			continue
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

// applyUniquenessByIntKey applies uniqueness logic using integer keys extracted from items.
func applyUniquenessByIntKey[T any](
	items []T,
	keyExtractor func(T) int,
) []T {
	seen := make(map[int]bool)
	result := make([]T, 0)
	for _, item := range items {
		key := keyExtractor(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

// applyUniquenessByCompositeKey applies uniqueness logic using composite string keys.
func applyUniquenessByCompositeKey[T any](
	items []T,
	keyExtractor func(T) string,
	separator string,
) []T {
	seen := make(map[string]bool)
	result := make([]T, 0)
	for _, item := range items {
		key := keyExtractor(item)
		if key == "" {
			continue
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

// ForEachIndividual iterates over all individuals in the graph, calling fn for each.
// Stops iteration if fn returns an error.
func ForEachIndividual(graph *Graph, fn func(*IndividualNode) error) error {
	if graph == nil {
		return fmt.Errorf("graph is nil")
	}
	allIndividuals := graph.GetAllIndividuals()
	for _, indiNode := range allIndividuals {
		if indiNode.Individual != nil {
			if err := fn(indiNode); err != nil {
				return err
			}
		}
	}
	return nil
}

// ForEachFamily iterates over all families in the graph, calling fn for each.
// Stops iteration if fn returns an error.
func ForEachFamily(graph *Graph, fn func(*FamilyNode) error) error {
	if graph == nil {
		return fmt.Errorf("graph is nil")
	}
	allFamilies := graph.GetAllFamilies()
	for _, famNode := range allFamilies {
		if famNode.Family != nil {
			if err := fn(famNode); err != nil {
				return err
			}
		}
	}
	return nil
}

// CollectIndividuals collects all individuals matching the filter.
func CollectIndividuals(graph *Graph, filter func(*IndividualNode) bool) []*IndividualNode {
	if graph == nil {
		return nil
	}
	allIndividuals := graph.GetAllIndividuals()
	result := make([]*IndividualNode, 0)
	for _, indiNode := range allIndividuals {
		if indiNode.Individual != nil && filter(indiNode) {
			result = append(result, indiNode)
		}
	}
	return result
}

// CollectFamilies collects all families matching the filter.
func CollectFamilies(graph *Graph, filter func(*FamilyNode) bool) []*FamilyNode {
	if graph == nil {
		return nil
	}
	allFamilies := graph.GetAllFamilies()
	result := make([]*FamilyNode, 0)
	for _, famNode := range allFamilies {
		if famNode.Family != nil && filter(famNode) {
			result = append(result, famNode)
		}
	}
	return result
}

