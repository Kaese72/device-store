package mariadb

import (
	"context"
	"testing"

	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/restmodels"
)

// TestGetDevicesNameFilter exercises the "name" device filter added for the
// devices-table name column filter in the UI, covering both the exact-match
// and case-insensitive substring operators.
func TestGetDevicesNameFilter(t *testing.T) {
	persistence := setupTestDB(t)
	ctx := context.Background()

	kitchenId, _, err := persistence.PostDevice(ctx, ingestmodels.IngestDevice{BridgeIdentifier: "bridge-kitchen", AdapterId: 1})
	if err != nil {
		t.Fatalf("failed to seed kitchen device: %v", err)
	}
	if _, err := persistence.UpdateDeviceName(ctx, kitchenId, "Kitchen Lamp"); err != nil {
		t.Fatalf("failed to name kitchen device: %v", err)
	}

	bedroomId, _, err := persistence.PostDevice(ctx, ingestmodels.IngestDevice{BridgeIdentifier: "bridge-bedroom", AdapterId: 1})
	if err != nil {
		t.Fatalf("failed to seed bedroom device: %v", err)
	}
	if _, err := persistence.UpdateDeviceName(ctx, bedroomId, "Bedroom Lamp"); err != nil {
		t.Fatalf("failed to name bedroom device: %v", err)
	}

	tests := []struct {
		name      string
		operator  string
		value     string
		expectIds []int
	}{
		{
			name:      "text-eq matches exactly",
			operator:  "text-eq",
			value:     "Kitchen Lamp",
			expectIds: []int{kitchenId},
		},
		{
			// The devices.name column uses the schema's default (case-insensitive)
			// collation, so "=" matches regardless of case - same as every other
			// VARCHAR column in this table (e.g. bridgeIdentifier).
			name:      "text-eq is case-insensitive, matching column collation",
			operator:  "text-eq",
			value:     "kitchen lamp",
			expectIds: []int{kitchenId},
		},
		{
			name:      "text-contains matches a substring case-insensitively",
			operator:  "text-contains",
			value:     "lamp",
			expectIds: []int{kitchenId, bedroomId},
		},
		{
			name:      "text-contains narrows to one device",
			operator:  "text-contains",
			value:     "Kitchen",
			expectIds: []int{kitchenId},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, total, err := persistence.GetDevices(ctx, []restmodels.Filter{
				{Key: "name", Operator: tt.operator, Value: tt.value},
			}, restmodels.Pagination{})
			if err != nil {
				t.Fatalf("GetDevices returned an error: %v", err)
			}
			if total != len(tt.expectIds) {
				t.Fatalf("total = %d, expected %d", total, len(tt.expectIds))
			}
			gotIds := make([]int, len(devices))
			for i, d := range devices {
				gotIds[i] = d.ID
			}
			for _, expected := range tt.expectIds {
				found := false
				for _, got := range gotIds {
					if got == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected device id %d in results, got %v", expected, gotIds)
				}
			}
		})
	}
}
