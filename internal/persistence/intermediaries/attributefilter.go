package intermediaries

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kaese72/huemie-lib/query"
	"github.com/danielgtaylor/huma/v2"
)

const attributeFilterKeyPrefix = "attribute."

type attributeOperator struct {
	column     string // the deviceAttributes column to compare against
	comparison string // the SQL comparison operator to use
	approx     bool   // whether to compare with a tolerance instead of exact equality
	contains   bool   // whether to do a case-insensitive substring match instead of equality
}

// attributeOperators defines the type-prefixed operators available when
// filtering on device attributes. The prefix determines which of the three
// typed columns (boolean, numeric, text) is compared against, since an
// attribute's type isn't known from its name alone. This is device-store's
// own schema knowledge, unlike the generic column filters in mariadb.go
// which are built from huemie-lib/query's reusable operator builders - an
// attribute filter instead needs to turn into an EXISTS subquery keyed on
// both the attribute's name and its typed value column, which those
// builders aren't shaped for.
var attributeOperators = map[string]attributeOperator{
	"bool-eq":       {column: "booleanValue", comparison: "="},
	"numeric-eq":    {column: "numericValue", comparison: "="},
	"numeric-aeq":   {column: "numericValue", comparison: "=", approx: true},
	"numeric-lt":    {column: "numericValue", comparison: "<"},
	"numeric-gt":    {column: "numericValue", comparison: ">"},
	"text-eq":       {column: "textValue", comparison: "="},
	"text-contains": {column: "textValue", contains: true},
}

// IsAttributeFilterKey reports whether a filter field targets a device
// attribute, i.e. is of the form "attribute.<name>".
func IsAttributeFilterKey(field string) bool {
	return strings.HasPrefix(field, attributeFilterKeyPrefix)
}

// SplitAttributeFilters separates filters targeting device attributes
// (fields of the form "attribute.<name>") from all other filters, so the
// two groups can be translated to SQL separately.
func SplitAttributeFilters(filters []query.Filter) (attributeFilters, otherFilters []query.Filter) {
	for _, filter := range filters {
		if IsAttributeFilterKey(filter.Field) {
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
func TranslateAttributeFiltersToQueryFragments(filters []query.Filter, deviceIdColumn string) ([]string, []any, error) {
	var fragments []string
	var values []any
	for _, filter := range filters {
		name := strings.TrimPrefix(filter.Field, attributeFilterKeyPrefix)
		if name == "" {
			return nil, nil, huma.Error400BadRequest(fmt.Sprintf("attribute filter field %q is missing an attribute name", filter.Field))
		}
		op, ok := attributeOperators[filter.Operator]
		if !ok {
			return nil, nil, huma.Error400BadRequest(fmt.Sprintf("may not filter on attribute %q with operator %q", name, filter.Operator))
		}

		var comparisonValue any
		switch op.column {
		case "booleanValue":
			if filter.Value != "true" && filter.Value != "false" {
				return nil, nil, huma.Error400BadRequest(fmt.Sprintf("attribute filter value %q for %q must be \"true\" or \"false\"", filter.Value, filter.Field))
			}
			comparisonValue = "0"
			if filter.Value == "true" {
				comparisonValue = "1"
			}
		case "numericValue":
			asFloat, err := strconv.ParseFloat(filter.Value, 64)
			if err != nil {
				return nil, nil, huma.Error400BadRequest(fmt.Sprintf("attribute filter value %q for %q must be numeric", filter.Value, filter.Field))
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
			values = append(values, name, comparisonValue, query.NumericApproxEpsilon)
			continue
		}
		if op.contains {
			pattern := "%" + query.LikeEscaper.Replace(comparisonValue.(string)) + "%"
			fragments = append(fragments, fmt.Sprintf(
				"EXISTS (SELECT 1 FROM deviceAttributes WHERE deviceAttributes.deviceId = %s AND deviceAttributes.name = ? AND LOWER(deviceAttributes.%s) LIKE LOWER(?) ESCAPE '\\\\')",
				deviceIdColumn, op.column,
			))
			values = append(values, name, pattern)
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
