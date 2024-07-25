package intermediaries

import (
	"fmt"
	"regexp"
)

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

func TranslateFiltersToQueryFragments(filters []Filter, filterMap map[string]map[string]func(string) (string, []string)) ([]string, []interface{}, error) {
	var whereClauses []string
	var values []interface{}
	for _, filter := range filters {
		if filterFuncMap, ok := filterMap[filter.Key]; ok {
			if filterFunc, ok := filterFuncMap[filter.Operator]; ok {
				whereClause, filterValues := filterFunc(filter.Value)
				whereClauses = append(whereClauses, whereClause)
				for _, filterValue := range filterValues {
					values = append(values, filterValue)
				}

			} else {
				return nil, nil, fmt.Errorf("may not filter with operator, %s, on attribute, %s", filter.Operator, filter.Key)

			}
		} else {
			return nil, nil, fmt.Errorf("may not filter on attribute, %s", filter.Key)
		}
	}
	return whereClauses, values, nil
}
