package mariadb

import (
	"testing"

	"github.com/Kaese72/device-store/ingestmodels"
)

func TestEqualRest(t *testing.T) {
	tests := []struct {
		name     string
		dbAttr   dbAttribute
		restAttr ingestmodels.Attribute
		expected bool
	}{
		{
			name: "Equal attributes",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				BooleanValue: toDbBoolean(ptrBool(true)),
				NumericValue: ptrFloat32(25.5),
				TextValue:    ptrString("hot"),
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Boolean: ptrBool(true),
				Numeric: ptrFloat32(25.5),
				Text:    ptrString("hot"),
			},
			expected: true,
		},
		{
			name: "Different names",
			dbAttr: dbAttribute{
				Name: "humidity",
			},
			restAttr: ingestmodels.Attribute{
				Name: "temperature",
			},
			expected: false,
		},
		{
			name: "Different boolean values",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				BooleanValue: toDbBoolean(ptrBool(false)),
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Boolean: ptrBool(true),
			},
			expected: false,
		},
		{
			name: "boolean nil and not nil",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				BooleanValue: nil,
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Boolean: ptrBool(true),
			},
			expected: false,
		},
		{
			name: "boolean not nil and nil",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				BooleanValue: toDbBoolean(ptrBool(true)),
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Boolean: nil,
			},
			expected: false,
		},
		{
			name: "boolean nil and true",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				BooleanValue: nil,
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Boolean: ptrBool(true),
			},
			expected: false,
		},
		{
			name: "Different numeric values",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				NumericValue: ptrFloat32(20.0),
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Numeric: ptrFloat32(25.0),
			},
			expected: false,
		},
		{
			name: "Numeric nil and not nil",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				NumericValue: ptrFloat32(20.0),
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Numeric: nil,
			},
			expected: false,
		},
		{
			name: "Different text values",
			dbAttr: dbAttribute{
				Name:      "irrelevant",
				TextValue: ptrString("cold"),
			},
			restAttr: ingestmodels.Attribute{
				Name: "irrelevant",
				Text: ptrString("hot"),
			},
			expected: false,
		},
		{
			name: "Nil boolean values",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				BooleanValue: nil,
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Boolean: nil,
			},
			expected: true,
		},
		{
			name: "Nil numeric values",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				NumericValue: nil,
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Numeric: nil,
			},
			expected: true,
		},
		{
			name: "Nil text values",
			dbAttr: dbAttribute{
				Name:      "irrelevant",
				TextValue: nil,
			},
			restAttr: ingestmodels.Attribute{
				Name: "irrelevant",
				Text: nil,
			},
			expected: true,
		},
		{
			name: "Text nil and not nil",
			dbAttr: dbAttribute{
				Name:      "irrelevant",
				TextValue: nil,
			},
			restAttr: ingestmodels.Attribute{
				Name: "irrelevant",
				Text: ptrString("hot"),
			},
			expected: false,
		},
		{
			name: "first OK but second not",
			dbAttr: dbAttribute{
				Name:         "irrelevant",
				TextValue:    nil,
				BooleanValue: toDbBoolean(ptrBool(true)),
			},
			restAttr: ingestmodels.Attribute{
				Name:    "irrelevant",
				Text:    nil,
				Boolean: ptrBool(false),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dbAttr.EqualRest(tt.restAttr)
			if result != tt.expected {
				t.Errorf("EqualRest() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrFloat32(f float32) *float32 {
	return &f
}

func ptrString(s string) *string {
	return &s
}
