package restmodels

import (
	"encoding/json"

	"github.com/danielgtaylor/huma/v2"
)

// this file describes how to use filters
// The filter is on the format "key[operator]=value"
// The key is the attribute to filter on
// The operator is the operation to perform on the attribute
// The value is the value to compare the attribute to
// Example: "name[eq]=John" will filter on the name attribute and only return objects where the name is "John".
// The operators and attributes which can be used for filtering is defined for each
// attribute closer to the database layer.

type Filter struct {
	Operator string `json:"op"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

func ParseQueryIntoFilters(filterString string) ([]Filter, error) {
	// The filter string is a JSON blob we can decode into the Filters type
	filters := []Filter{}
	if filterString == "" {
		return filters, nil
	}
	// json unmarshal
	err := json.Unmarshal([]byte(filterString), &filters)
	if err != nil {
		return filters, huma.Error404NotFound(err.Error())
	}
	return filters, nil
}
