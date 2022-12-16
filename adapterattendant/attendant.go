package adapterattendant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kaese72/adapter-attendant/rest/models"
	"github.com/Kaese72/device-store/config"
)

type Attendant interface {
	GetAdapter(string) (models.Adapter, error)
}

func NewAdapterAttendant(config config.AdapterAttendantConfig) Attendant {
	return attendantClient{
		URL:   config.URL,
		cache: map[string]AdapterCache{},
	}
}

type AdapterCache struct {
	LastUpdate time.Time
	models.Adapter
}

type attendantClient struct {
	URL   string
	cache map[string]AdapterCache
}

func (client attendantClient) GetAdapter(adapterName string) (models.Adapter, error) {
	if cached, ok := client.cache[adapterName]; ok && cached.LastUpdate.After(time.Now().Add(-1*time.Hour)) {
		return cached.Adapter, nil
	}
	resp, err := http.Get(fmt.Sprintf("%s/rest/v0/adapters/%s", client.URL, adapterName))
	if err != nil {
		return models.Adapter{}, err
	}
	defer resp.Body.Close()
	adapter := models.Adapter{}
	err = json.NewDecoder(resp.Body).Decode(&adapter)
	if err != nil {
		return adapter, err
	}
	client.cache[adapterName] = AdapterCache{
		LastUpdate: time.Now(),
		Adapter:    adapter,
	}
	return adapter, err
}
