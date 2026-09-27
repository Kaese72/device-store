package intermediaries

import (
	"testing"

	"github.com/Kaese72/huemie-lib/query"
)

func TestSplitAttributeFilters(t *testing.T) {
	filters := []query.Filter{
		{Field: "id", Operator: "eq", Value: "1"},
		{Field: "attribute.active", Operator: "bool-eq", Value: "true"},
		{Field: "bridge-identifier", Operator: "eq", Value: "abc"},
		{Field: "attribute.color-ct", Operator: "numeric-aeq", Value: "0.123"},
	}
	attributeFilters, otherFilters := SplitAttributeFilters(filters)
	if len(attributeFilters) != 2 {
		t.Fatalf("expected 2 attribute filters, got %d", len(attributeFilters))
	}
	if len(otherFilters) != 2 {
		t.Fatalf("expected 2 other filters, got %d", len(otherFilters))
	}
	if attributeFilters[0].Field != "attribute.active" || attributeFilters[1].Field != "attribute.color-ct" {
		t.Errorf("unexpected attribute filters: %+v", attributeFilters)
	}
	if otherFilters[0].Field != "id" || otherFilters[1].Field != "bridge-identifier" {
		t.Errorf("unexpected other filters: %+v", otherFilters)
	}
}

func TestTranslateAttributeFiltersToQueryFragments(t *testing.T) {
	tests := []struct {
		name        string
		filter      query.Filter
		expectError bool
		expectValue []any
	}{
		{
			name:        "boolean eq true",
			filter:      query.Filter{Field: "attribute.active", Operator: "bool-eq", Value: "true"},
			expectValue: []any{"active", "1"},
		},
		{
			name:        "boolean eq false",
			filter:      query.Filter{Field: "attribute.active", Operator: "bool-eq", Value: "false"},
			expectValue: []any{"active", "0"},
		},
		{
			name:        "boolean eq invalid value",
			filter:      query.Filter{Field: "attribute.active", Operator: "bool-eq", Value: "yes"},
			expectError: true,
		},
		{
			name:        "numeric eq",
			filter:      query.Filter{Field: "attribute.color-ct", Operator: "numeric-eq", Value: "1.5"},
			expectValue: []any{"color-ct", 1.5},
		},
		{
			name:        "numeric eq invalid value",
			filter:      query.Filter{Field: "attribute.color-ct", Operator: "numeric-eq", Value: "not-a-number"},
			expectError: true,
		},
		{
			name:        "numeric approx eq",
			filter:      query.Filter{Field: "attribute.color-ct", Operator: "numeric-aeq", Value: "0.123"},
			expectValue: []any{"color-ct", 0.123, query.NumericApproxEpsilon},
		},
		{
			name:        "numeric lt",
			filter:      query.Filter{Field: "attribute.color-ct", Operator: "numeric-lt", Value: "1.6"},
			expectValue: []any{"color-ct", 1.6},
		},
		{
			name:        "numeric gt",
			filter:      query.Filter{Field: "attribute.color-ct", Operator: "numeric-gt", Value: "1.6"},
			expectValue: []any{"color-ct", 1.6},
		},
		{
			name:        "text eq",
			filter:      query.Filter{Field: "attribute.name", Operator: "text-eq", Value: "hello"},
			expectValue: []any{"name", "hello"},
		},
		{
			name:        "text contains",
			filter:      query.Filter{Field: "attribute.name", Operator: "text-contains", Value: "ell"},
			expectValue: []any{"name", "%ell%"},
		},
		{
			name:        "text contains escapes LIKE wildcards",
			filter:      query.Filter{Field: "attribute.name", Operator: "text-contains", Value: "50%_off"},
			expectValue: []any{"name", `%50\%\_off%`},
		},
		{
			name:        "unsupported operator for attributes",
			filter:      query.Filter{Field: "attribute.active", Operator: "bool-lt", Value: "true"},
			expectError: true,
		},
		{
			name:        "unknown operator",
			filter:      query.Filter{Field: "attribute.active", Operator: "eq", Value: "true"},
			expectError: true,
		},
		{
			name:        "missing attribute name",
			filter:      query.Filter{Field: "attribute.", Operator: "bool-eq", Value: "true"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fragments, values, err := TranslateAttributeFiltersToQueryFragments([]query.Filter{tt.filter}, "devices.id")
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fragments) != 1 {
				t.Fatalf("expected 1 fragment, got %d", len(fragments))
			}
			if len(values) != len(tt.expectValue) {
				t.Fatalf("expected %d values, got %d: %+v", len(tt.expectValue), len(values), values)
			}
			for i, v := range tt.expectValue {
				if values[i] != v {
					t.Errorf("value %d: expected %v (%T), got %v (%T)", i, v, v, values[i], values[i])
				}
			}
		})
	}
}
