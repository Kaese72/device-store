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

func TestValidateTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		// Valid full timestamp format (YYYY-MM-DD HH:MM:SS)
		{
			name:        "Valid full timestamp",
			input:       "2025-08-24 14:30:25",
			expectError: false,
		},
		{
			name:        "Valid full timestamp - midnight",
			input:       "2025-01-01 00:00:00",
			expectError: false,
		},
		{
			name:        "Valid full timestamp - end of day",
			input:       "2025-12-31 23:59:59",
			expectError: false,
		},
		// Valid date format (YYYY-MM-DD)
		{
			name:        "Valid date format",
			input:       "2025-08-24",
			expectError: false,
		},
		{
			name:        "Valid date format - leap year",
			input:       "2024-02-29",
			expectError: false,
		},
		{
			name:        "Valid date format - first day of year",
			input:       "2025-01-01",
			expectError: false,
		},
		{
			name:        "Valid date format - last day of year",
			input:       "2025-12-31",
			expectError: false,
		},
		// Invalid formats
		{
			name:        "Invalid format - missing time components",
			input:       "2025-08-24 14:30",
			expectError: true,
		},
		{
			name:        "Invalid format - extra seconds precision",
			input:       "2025-08-24 14:30:25.123",
			expectError: true,
		},
		{
			name:        "Invalid format - wrong separator",
			input:       "2025/08/24 14:30:25",
			expectError: true,
		},
		{
			name:        "Invalid format - missing leading zeros",
			input:       "2025-8-24 14:30:25",
			expectError: true,
		},
		{
			name:        "Invalid format - single digit hour",
			input:       "2025-08-24 4:30:25",
			expectError: true,
		},
		{
			name:        "Invalid format - reversed date",
			input:       "24-08-2025",
			expectError: true,
		},
		{
			name:        "Invalid format - empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "Invalid format - random text",
			input:       "not a timestamp",
			expectError: true,
		},
		// Invalid dates that match format but are not valid dates
		{
			name:        "Invalid date - month 13",
			input:       "2025-13-01",
			expectError: true,
		},
		{
			name:        "Invalid date - day 32",
			input:       "2025-01-32",
			expectError: true,
		},
		{
			name:        "Invalid date - February 30",
			input:       "2025-02-30",
			expectError: true,
		},
		{
			name:        "Invalid date - February 29 non-leap year",
			input:       "2025-02-29",
			expectError: true,
		},
		{
			name:        "Invalid time - hour 25",
			input:       "2025-08-24 25:30:25",
			expectError: true,
		},
		{
			name:        "Invalid time - minute 60",
			input:       "2025-08-24 14:60:25",
			expectError: true,
		},
		{
			name:        "Invalid time - second 60",
			input:       "2025-08-24 14:30:60",
			expectError: true,
		},
		// Edge cases
		{
			name:        "Valid - minimum date",
			input:       "0001-01-01",
			expectError: false,
		},
		{
			name:        "Valid - year 9999",
			input:       "9999-12-31",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimestamp(tt.input)
			hasError := err != nil
			if hasError != tt.expectError {
				if tt.expectError {
					t.Errorf("validateTimestamp(%q) expected error but got none", tt.input)
				} else {
					t.Errorf("validateTimestamp(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}
