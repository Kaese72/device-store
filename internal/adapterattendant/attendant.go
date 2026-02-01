package adapterattendant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kaese72/device-store/internal/config"
	"go.elastic.co/apm/module/apmhttp/v2"
	"golang.org/x/net/context/ctxhttp"
)

func NewAdapterTrigger(config config.AdapterAttendantConfig) AdapterTriggerClient {
	return AdapterTriggerClient{
		URL:   config.URL,
		cache: map[int]AdapterAddressCache{},
	}
}

type AdapterAddressCache struct {
	LastUpdate time.Time
	Address    string
}

type AdapterTriggerClient struct {
	URL   string
	cache map[int]AdapterAddressCache
}

var tracingClient = apmhttp.WrapClient(http.DefaultClient)

func (client AdapterTriggerClient) GetAdapterAddress(ctx context.Context, adapterId int) (string, error) {
	if cached, ok := client.cache[adapterId]; ok && cached.LastUpdate.After(time.Now().Add(-1*time.Hour)) {
		return cached.Address, nil
	}
	resp, err := ctxhttp.Get(ctx, tracingClient, fmt.Sprintf("%s/adapter-attendant/v1/adapters/%d/address", client.URL, adapterId))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	response := struct {
		Address string `json:"address"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}
	client.cache[adapterId] = AdapterAddressCache{
		LastUpdate: time.Now(),
		Address:    response.Address,
	}
	return response.Address, nil
}
