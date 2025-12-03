package util

import (
	"fmt"
	"strings"
)

// SortCondition represents a single sorting condition with field and direction.
type SortCondition struct {
	Field     string
	Direction string // "asc" or "desc"
}

// ParseSortBy parses the sort_by string into a list of SQL sort conditions.
// Example: "name,created_at:desc" -> ["name asc", "created_at desc"]
func ParseSortBy(sortBy string) []string {
	if sortBy == "" {
		return nil
	}

	fields := strings.Split(sortBy, ",")
	sortConditions := make([]string, 0, len(fields))

	for _, field := range fields {
		trimmedField := strings.TrimSpace(field)
		parts := strings.SplitN(trimmedField, ":", 2)
		fieldName := strings.TrimSpace(parts[0])
		direction := "asc" // Default direction

		if len(parts) > 1 {
			dir := strings.ToLower(strings.TrimSpace(parts[1]))
			if dir == "desc" {
				direction = "desc"
			}
		}

		if fieldName != "" {
			sortConditions = append(sortConditions, fmt.Sprintf("%s %s", fieldName, direction))
		}
	}

	return sortConditions
}

// GetOrderExpr constructs the order expression for a BUN query.
// It takes a slice of strings, where each string represents a sorting condition
// (e.g., "name asc", "created_at desc"). If the input slice is empty, it returns
// a default sort order.
//
// Example:
//
//	orderExprs := []string{"name asc", "created_at desc"}
//	expr := GetOrderExpr(orderExprs, "created_at DESC")
//	// expr will be "name asc, created_at desc"
//
//	orderExprs := []string{}
//	expr := GetOrderExpr(orderExprs, "created_at DESC")
//	// expr will be "created_at DESC"
func GetOrderExpr(orderExprs []string, defaultSort string) string {
	if len(orderExprs) == 0 {
		return defaultSort
	}
	return strings.Join(orderExprs, ", ")
}
