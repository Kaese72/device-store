package intermediaries_test

import (
	"reflect"
	"testing"

	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
)

func TestParseQueryIntoFilter(t *testing.T) {
	tests := []struct {
		name            string
		queryParameters map[string][]string
		expectedFilters []intermediaries.Filter
	}{
		{
			name:            "Empty query parameters",
			queryParameters: map[string][]string{},
			expectedFilters: []intermediaries.Filter{},
		},
		{
			name:            "Single query parameter",
			queryParameters: map[string][]string{"key[operator]": {"value"}},
			expectedFilters: []intermediaries.Filter{
				{
					Operator: "operator",
					Key:      "key",
					Value:    "value",
				},
			},
		},
		{
			name:            "Single query parameter",
			queryParameters: map[string][]string{"": {"=value"}},
			expectedFilters: []intermediaries.Filter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intermediaries.ParseQueryIntoFilters(tt.queryParameters); !reflect.DeepEqual(got, tt.expectedFilters) {
				t.Errorf("ParseQueryIntoFilters() = %v, want %v", got, tt.expectedFilters)
			}
		})
	}
}
