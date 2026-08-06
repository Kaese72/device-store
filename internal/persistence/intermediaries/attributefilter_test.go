package intermediaries

import (
	"testing"

	"github.com/Kaese72/device-store/restmodels"
)

func TestSplitAttributeFilters(t *testing.T) {
	filters := []restmodels.Filter{
		{Key: "id", Operator: "eq", Value: "1"},
		{Key: "attribute.active", Operator: "bool-eq", Value: "true"},
		{Key: "bridge-identifier", Operator: "eq", Value: "abc"},
		{Key: "attribute.color-ct", Operator: "numeric-aeq", Value: "0.123"},
	}
	attributeFilters, otherFilters := SplitAttributeFilters(filters)
	if len(attributeFilters) != 2 {
		t.Fatalf("expected 2 attribute filters, got %d", len(attributeFilters))
	}
	if len(otherFilters) != 2 {
		t.Fatalf("expected 2 other filters, got %d", len(otherFilters))
	}
	if attributeFilters[0].Key != "attribute.active" || attributeFilters[1].Key != "attribute.color-ct" {
		t.Errorf("unexpected attribute filters: %+v", attributeFilters)
	}
	if otherFilters[0].Key != "id" || otherFilters[1].Key != "bridge-identifier" {
		t.Errorf("unexpected other filters: %+v", otherFilters)
	}
}

func TestTranslateAttributeFiltersToQueryFragments(t *testing.T) {
	tests := []struct {
		name        string
		filter      restmodels.Filter
		expectError bool
		expectValue []any
	}{
		{
			name:        "boolean eq true",
			filter:      restmodels.Filter{Key: "attribute.active", Operator: "bool-eq", Value: "true"},
			expectValue: []any{"active", "1"},
		},
		{
			name:        "boolean eq false",
			filter:      restmodels.Filter{Key: "attribute.active", Operator: "bool-eq", Value: "false"},
			expectValue: []any{"active", "0"},
		},
		{
			name:        "boolean eq invalid value",
			filter:      restmodels.Filter{Key: "attribute.active", Operator: "bool-eq", Value: "yes"},
			expectError: true,
		},
		{
			name:        "numeric eq",
			filter:      restmodels.Filter{Key: "attribute.color-ct", Operator: "numeric-eq", Value: "1.5"},
			expectValue: []any{"color-ct", 1.5},
		},
		{
			name:        "numeric eq invalid value",
			filter:      restmodels.Filter{Key: "attribute.color-ct", Operator: "numeric-eq", Value: "not-a-number"},
			expectError: true,
		},
		{
			name:        "numeric approx eq",
			filter:      restmodels.Filter{Key: "attribute.color-ct", Operator: "numeric-aeq", Value: "0.123"},
			expectValue: []any{"color-ct", 0.123, attributeApproxEpsilon},
		},
		{
			name:        "numeric lt",
			filter:      restmodels.Filter{Key: "attribute.color-ct", Operator: "numeric-lt", Value: "1.6"},
			expectValue: []any{"color-ct", 1.6},
		},
		{
			name:        "numeric gt",
			filter:      restmodels.Filter{Key: "attribute.color-ct", Operator: "numeric-gt", Value: "1.6"},
			expectValue: []any{"color-ct", 1.6},
		},
		{
			name:        "text eq",
			filter:      restmodels.Filter{Key: "attribute.name", Operator: "text-eq", Value: "hello"},
			expectValue: []any{"name", "hello"},
		},
		{
			name:        "unsupported operator for attributes",
			filter:      restmodels.Filter{Key: "attribute.active", Operator: "bool-lt", Value: "true"},
			expectError: true,
		},
		{
			name:        "unknown operator",
			filter:      restmodels.Filter{Key: "attribute.active", Operator: "eq", Value: "true"},
			expectError: true,
		},
		{
			name:        "missing attribute name",
			filter:      restmodels.Filter{Key: "attribute.", Operator: "bool-eq", Value: "true"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fragments, values, err := TranslateAttributeFiltersToQueryFragments([]restmodels.Filter{tt.filter}, "devices.id")
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
