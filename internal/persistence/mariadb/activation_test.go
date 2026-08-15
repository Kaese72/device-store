package mariadb

import (
	"context"
	"testing"

	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/restmodels"
)

// TestGetDeviceCapabilityForActivation seeds a device whose name equals the
// capability's own name, and a device name that also collides with the
// capability being looked up. If any query in this path references an
// unqualified "name" column, it either errors with "ambiguous column" or
// silently resolves to the wrong table's value - both would fail here.
func TestGetDeviceCapabilityForActivation(t *testing.T) {
	persistence := setupTestDB(t)
	ctx := context.Background()

	deviceId, _, err := persistence.PostDevice(ctx, ingestmodels.IngestDevice{
		BridgeIdentifier: "bridge-1",
		AdapterId:        1,
		Capabilities: []ingestmodels.IngestDeviceCapability{
			{Name: "activate"},
		},
	})
	if err != nil {
		t.Fatalf("failed to seed device: %v", err)
	}

	// Give the device the same name as the capability we are about to look
	// up. This is exactly the scenario that made "name" ambiguous once
	// devices gained their own name column.
	if _, err := persistence.UpdateDeviceName(ctx, deviceId, "activate"); err != nil {
		t.Fatalf("failed to set device name: %v", err)
	}

	capability, err := persistence.GetDeviceCapabilityForActivation(ctx, deviceId, "activate")
	if err != nil {
		t.Fatalf("GetDeviceCapabilityForActivation returned an error: %v", err)
	}
	if capability.Name != "activate" {
		t.Errorf("capability.Name = %q, expected %q", capability.Name, "activate")
	}
	if capability.BridgeIdentifier != "bridge-1" {
		t.Errorf("capability.BridgeIdentifier = %q, expected %q", capability.BridgeIdentifier, "bridge-1")
	}
	if capability.AdapterId != 1 {
		t.Errorf("capability.AdapterId = %d, expected %d", capability.AdapterId, 1)
	}
}

// TestGetGroupCapabilityForActivation is the group-side counterpart, kept
// alongside the device test since both queries share the same join shape.
func TestGetGroupCapabilityForActivation(t *testing.T) {
	persistence := setupTestDB(t)
	ctx := context.Background()

	err := persistence.PostGroup(ctx, ingestmodels.IngestGroup{
		Name:             "activate",
		BridgeIdentifier: "bridge-1",
		AdapterId:        1,
		Capabilities: []ingestmodels.IngestGroupCapability{
			{Name: "activate"},
		},
	})
	if err != nil {
		t.Fatalf("failed to seed group: %v", err)
	}

	groups, _, err := persistence.GetGroups(ctx, nil, restmodels.Pagination{})
	if err != nil {
		t.Fatalf("failed to look up seeded group: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected exactly one group, got %d", len(groups))
	}

	capability, err := persistence.GetGroupCapabilityForActivation(ctx, groups[0].ID, "activate")
	if err != nil {
		t.Fatalf("GetGroupCapabilityForActivation returned an error: %v", err)
	}
	if capability.Name != "activate" {
		t.Errorf("capability.Name = %q, expected %q", capability.Name, "activate")
	}
}
