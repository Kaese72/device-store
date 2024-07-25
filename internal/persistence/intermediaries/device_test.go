package intermediaries_test

import (
	"reflect"
	"testing"

	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/device-store/rest/models"
)

func TestDeviceFilterPresent(t *testing.T) {
	device := models.Device{}
	for filterKey := range intermediaries.DeviceFilters {
		t.Run(filterKey, func(t *testing.T) {
			nFields := reflect.TypeOf(device).NumField()
			found := false
			for i := 0; i < nFields; i++ {
				field := reflect.TypeOf(device).Field(i)
				if field.Tag.Get("json") == filterKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("filter key %s not found in device struct", filterKey)
			}
		})
	}
}
