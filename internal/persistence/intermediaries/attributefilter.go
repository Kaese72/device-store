package intermediaries

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kaese72/device-store/restmodels"
	"github.com/danielgtaylor/huma/v2"
)

const attributeFilterKeyPrefix = "attribute."

// attributeApproxEpsilon mirrors the tolerance used elsewhere when comparing
// attribute values for equality (see dbAttribute.EqualRest), since attribute
// values are floating point and prone to rounding error.
const attributeApproxEpsilon = 0.001

type attributeOperator struct {
	column     string // the deviceAttributes column to compare against
	comparison string // the SQL comparison operator to use
	approx     bool   // whether to compare with a tolerance instead of exact equality
}

// attributeOperators defines the type-prefixed operators available when
// filtering on device attributes. The prefix determines which of the three
// typed columns (boolean, numeric, text) is compared against, since an
// attribute's type isn't known from its name alone.
var attributeOperators = map[string]attributeOperator{
	"bool-eq":     {column: "booleanValue", comparison: "="},
	"numeric-eq":  {column: "numericValue", comparison: "="},
	"numeric-aeq": {column: "numericValue", comparison: "=", approx: true},
	"numeric-lt":  {column: "numericValue", comparison: "<"},
	"numeric-gt":  {column: "numericValue", comparison: ">"},
	"text-eq":     {column: "textValue", comparison: "="},
}

// IsAttributeFilterKey reports whether a filter key targets a device
// attribute, i.e. is of the form "attribute.<name>".
func IsAttributeFilterKey(key string) bool {
	return strings.HasPrefix(key, attributeFilterKeyPrefix)
}

// SplitAttributeFilters separates filters targeting device attributes (keys
// of the form "attribute.<name>") from all other filters, so the two groups
// can be translated to SQL separately.
func SplitAttributeFilters(filters []restmodels.Filter) (attributeFilters, otherFilters []restmodels.Filter) {
	for _, filter := range filters {
		if IsAttributeFilterKey(filter.Key) {
			attributeFilters = append(attributeFilters, filter)
		} else {
			otherFilters = append(otherFilters, filter)
		}
	}
	return attributeFilters, otherFilters
}

// TranslateAttributeFiltersToQueryFragments turns "attribute.<name>" filters
// into SQL fragments matched against the deviceAttributes table, scoped to
// the given devices-table id column (e.g. "devices.id"). A device with no
// matching attribute row, or with the attribute stored under a different
// type than the operator implies, does not match.
func TranslateAttributeFiltersToQueryFragments(filters []restmodels.Filter, deviceIdColumn string) ([]string, []any, error) {
	var fragments []string
	var values []any
	for _, filter := range filters {
		name := strings.TrimPrefix(filter.Key, attributeFilterKeyPrefix)
		if name == "" {
			return nil, nil, huma.Error400BadRequest(fmt.Sprintf("attribute filter key %q is missing an attribute name", filter.Key))
		}
		op, ok := attributeOperators[filter.Operator]
		if !ok {
			return nil, nil, huma.Error400BadRequest(fmt.Sprintf("may not filter on attribute %q with operator %q", name, filter.Operator))
		}

		var comparisonValue any
		switch op.column {
		case "booleanValue":
			if filter.Value != "true" && filter.Value != "false" {
				return nil, nil, huma.Error400BadRequest(fmt.Sprintf("attribute filter value %q for %q must be \"true\" or \"false\"", filter.Value, filter.Key))
			}
			comparisonValue = "0"
			if filter.Value == "true" {
				comparisonValue = "1"
			}
		case "numericValue":
			asFloat, err := strconv.ParseFloat(filter.Value, 64)
			if err != nil {
				return nil, nil, huma.Error400BadRequest(fmt.Sprintf("attribute filter value %q for %q must be numeric", filter.Value, filter.Key))
			}
			comparisonValue = asFloat
		default: // textValue
			comparisonValue = filter.Value
		}

		if op.approx {
			fragments = append(fragments, fmt.Sprintf(
				"EXISTS (SELECT 1 FROM deviceAttributes WHERE deviceAttributes.deviceId = %s AND deviceAttributes.name = ? AND ABS(deviceAttributes.%s - ?) < ?)",
				deviceIdColumn, op.column,
			))
			values = append(values, name, comparisonValue, attributeApproxEpsilon)
			continue
		}
		fragments = append(fragments, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM deviceAttributes WHERE deviceAttributes.deviceId = %s AND deviceAttributes.name = ? AND deviceAttributes.%s %s ?)",
			deviceIdColumn, op.column, op.comparison,
		))
		values = append(values, name, comparisonValue)
	}
	return fragments, values, nil
}
