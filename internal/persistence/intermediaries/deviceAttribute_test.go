package intermediaries_test

import (
	"reflect"
	"testing"

	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/device-store/rest/models"
)

func TestToRestModel(t *testing.T) {
	tests := []struct {
		name     string
		input    intermediaries.AttributeIntermediary
		expected models.Attribute
	}{
		{
			name: "all nil values",
			input: intermediaries.AttributeIntermediary{
				Name:     "name",
				DeviceId: "deviceId",
				BooleanX: nil,
				Numeric:  nil,
				Text:     nil,
			},
			expected: models.Attribute{
				Name:    "name",
				Boolean: nil,
				Numeric: nil,
				Text:    nil,
			},
		},
		{
			name: "all set to valid values",
			input: intermediaries.AttributeIntermediary{
				Name:     "name",
				DeviceId: "deviceId",
				BooleanX: &[]float32{1}[0],
				Numeric:  &[]float32{123}[0],
				Text:     &[]string{"yeet"}[0],
			},
			expected: models.Attribute{
				Name:    "name",
				Boolean: &[]bool{true}[0],
				Numeric: &[]float32{123}[0],
				Text:    &[]string{"yeet"}[0],
			},
		},
		{
			name: "all set to empty values (but set)",
			input: intermediaries.AttributeIntermediary{
				Name:     "name",
				DeviceId: "deviceId",
				BooleanX: &[]float32{0}[0],
				Numeric:  &[]float32{0}[0],
				Text:     &[]string{""}[0],
			},
			expected: models.Attribute{
				Name:    "name",
				Boolean: &[]bool{false}[0],
				Numeric: &[]float32{0}[0],
				Text:    &[]string{""}[0],
			},
		},
		{
			name: "boolean set weirdly",
			input: intermediaries.AttributeIntermediary{
				Name:     "name",
				DeviceId: "deviceId",
				BooleanX: &[]float32{1337}[0],
				Numeric:  nil,
				Text:     nil,
			},
			expected: models.Attribute{
				Name:    "name",
				Boolean: &[]bool{true}[0],
				Numeric: nil,
				Text:    nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.input.ToRestModel()
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}
