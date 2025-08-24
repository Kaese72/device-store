package intermediaries

import (
	"fmt"

	"github.com/Kaese72/device-store/restmodels"
)

func TranslateFiltersToQueryFragments(filters []restmodels.Filter, filterMap map[string]map[string]func(string) (string, []string, error)) ([]string, []any, error) {
	var whereClauses []string
	var values []interface{}
	for _, filter := range filters {
		if filterFuncMap, ok := filterMap[filter.Key]; ok {
			if filterFunc, ok := filterFuncMap[filter.Operator]; ok {
				whereClause, filterValues, err := filterFunc(filter.Value)
				if err != nil {
					return nil, nil, err
				}
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
