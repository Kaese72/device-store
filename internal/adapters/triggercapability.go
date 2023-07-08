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
	"github.com/Kaese72/device-store/internal/systemerrors"
	"github.com/Kaese72/device-store/rest/models"
	"go.elastic.co/apm/module/apmhttp"
	"golang.org/x/net/context/ctxhttp"
)

var tracingClient = apmhttp.WrapClient(http.DefaultClient)

func TriggerDeviceCapability(ctx context.Context, adapter adapterattendantmodels.Adapter, deviceID string, capabilityID string, capArg models.CapabilityArgs) systemerrors.SystemError {
	jsonEncoded, err := json.Marshal(capArg)
	if err != nil {
		return systemerrors.WrapSystemError(err, systemerrors.UserError)
	}

	adapterURL, err := url.Parse(adapter.Address)
	if err != nil {
		return systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}

	adapterURL.Path = fmt.Sprintf("devices/%s/capabilities/%s", deviceID, capabilityID)
	logging.Info("Triggering capability", ctx, map[string]interface{}{"capUri": adapterURL.String()})
	resp, err := ctxhttp.Post(ctx, tracingClient, adapterURL.String(), "application/json", bytes.NewBuffer(jsonEncoded))
	if err != nil {
		// FIXME What if there is interesting debug information in the response?
		// We should log it or incorporate it in the response message or something
		return systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	// It is the callers responsibility to Close the body reader
	// But there should not be anything of interest here at the moment
	defer resp.Body.Close()
	logging.Info("Capability triggered", ctx, map[string]interface{}{"rCode": strconv.Itoa(resp.StatusCode)})
	return nil
}
