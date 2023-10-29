package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	adapterattendantmodels "github.com/Kaese72/adapter-attendant/rest/models"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/rest/models"
	"github.com/Kaese72/huemie-lib/liberrors"
	"go.elastic.co/apm/module/apmhttp/v2"
	"golang.org/x/net/context/ctxhttp"
)

var tracingClient = apmhttp.WrapClient(http.DefaultClient)

func TriggerDeviceCapability(ctx context.Context, adapter adapterattendantmodels.Adapter, bridgeDeviceIdentifier string, capabilityID string, capArg models.CapabilityArgs) error {
	jsonEncoded, err := json.Marshal(capArg)
	if err != nil {
		return liberrors.NewApiError(liberrors.UserError, err)
	}

	adapterURL, err := url.Parse(adapter.Address)
	if err != nil {
		return liberrors.NewApiError(liberrors.InternalError, err)
	}

	adapterURL.Path = fmt.Sprintf("devices/%s/capabilities/%s", bridgeDeviceIdentifier, capabilityID)
	logging.Info("Triggering capability", ctx, map[string]interface{}{"capUri": adapterURL.String()})
	resp, err := ctxhttp.Post(ctx, tracingClient, adapterURL.String(), "application/json", bytes.NewBuffer(jsonEncoded))
	if err != nil {
		// FIXME What if there is interesting debug information in the response?
		// We should log it or incorporate it in the response message or something
		return liberrors.NewApiError(liberrors.InternalError, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logging.Info("Capability failed to trigger", ctx, map[string]interface{}{"rCode": strconv.Itoa(resp.StatusCode)})
		return liberrors.NewApiError(liberrors.InternalError, fmt.Errorf("not HTTP/200, %d", resp.StatusCode))
	}
	// It is the callers responsibility to Close the body reader
	// But there should not be anything of interest here at the moment
	logging.Info("Capability triggered", ctx, map[string]interface{}{"rCode": strconv.Itoa(resp.StatusCode)})
	return nil
}

func TriggerGroupCapability(ctx context.Context, adapter adapterattendantmodels.Adapter, groupId string, capabilityId string, capArg models.CapabilityArgs) error {
	jsonEncoded, err := json.Marshal(capArg)
	if err != nil {
		return liberrors.NewApiError(liberrors.UserError, err)
	}

	adapterURL, err := url.Parse(adapter.Address)
	if err != nil {
		return liberrors.NewApiError(liberrors.InternalError, err)
	}

	adapterURL.Path = fmt.Sprintf("groups/%s/capabilities/%s", groupId, capabilityId)
	logging.Info("Triggering capability", ctx, map[string]interface{}{"capUri": adapterURL.String()})
	resp, err := ctxhttp.Post(ctx, tracingClient, adapterURL.String(), "application/json", bytes.NewBuffer(jsonEncoded))
	if err != nil {
		// FIXME What if there is interesting debug information in the response?
		// We should log it or incorporate it in the response message or something
		return liberrors.NewApiError(liberrors.InternalError, err)
	}
	// It is the callers responsibility to Close the body reader
	// But there should not be anything of interest here at the moment
	defer resp.Body.Close()
	logging.Info("Capability triggered", ctx, map[string]interface{}{"rCode": strconv.Itoa(resp.StatusCode)})
	return nil
}
