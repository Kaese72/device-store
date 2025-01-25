package restmodels

import "regexp"

// this file describes how to use filters
// The filter is on the format "key[operator]=value"
// The key is the attribute to filter on
// The operator is the operation to perform on the attribute
// The value is the value to compare the attribute to
// Example: "name[eq]=John" will filter on the name attribute and only return objects where the name is "John".
// The operators and attributes which can be used for filtering is defined for each
// attribute closer to the database layer.

type Filter struct {
	Operator string
	Key      string
	Value    string
}

var queryKeyRegex = regexp.MustCompile(`^(.+)\[(.+)\]$`)

func ParseQueryIntoFilters(queryParameters map[string][]string) []Filter {
	// Each query parameter is on the format "key[operator]=value"
	filters := []Filter{}
	for queryKey, queryKeyValues := range queryParameters {
		found := queryKeyRegex.FindStringSubmatch(queryKey)
		if found == nil {
			// Not a valid query parameter... ignore
			continue
		}
		for _, queryParameter := range queryKeyValues {
			// Split the query parameter into key, operator and value

			filters = append(filters, Filter{Operator: found[2], Key: found[1], Value: queryParameter})
		}
	}
	return filters
}
