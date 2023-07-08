package adapterattendant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kaese72/adapter-attendant/rest/models"
	"github.com/Kaese72/device-store/internal/config"
	"go.elastic.co/apm/module/apmhttp/v2"
	"golang.org/x/net/context/ctxhttp"
)

type Attendant interface {
	GetAdapter(string, context.Context) (models.Adapter, error)
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

var tracingClient = apmhttp.WrapClient(http.DefaultClient)

func (client attendantClient) GetAdapter(adapterName string, ctx context.Context) (models.Adapter, error) {
	if cached, ok := client.cache[adapterName]; ok && cached.LastUpdate.After(time.Now().Add(-1*time.Hour)) {
		return cached.Adapter, nil
	}
	resp, err := ctxhttp.Get(ctx, tracingClient, fmt.Sprintf("%s/adapter-attendant/v0/adapters/%s", client.URL, adapterName))
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
